package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	monitorv1alpha1 "github.com/iam404/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/iam404/endpoint-monitoring-operator/internal/driver"
	"github.com/iam404/endpoint-monitoring-operator/internal/notifier"
)

// fakeDriver is a test double for driver.Driver.
type fakeDriver struct {
	result *driver.CheckResult
	err    error
	called bool
}

func (f *fakeDriver) Check() (*driver.CheckResult, error) {
	f.called = true
	return f.result, f.err
}
func (f *fakeDriver) GetEndpoint() string { return "http://fake" }
func (f *fakeDriver) GetType() string     { return "fake" }

// fakeNotifier is a test double for notifier.Notifier.
type fakeNotifier struct {
	gotStatus string
	gotMsg    string
	called    bool
	err       error
}

func (f *fakeNotifier) SendAlert(status, msg string) error {
	f.called = true
	f.gotStatus = status
	f.gotMsg = msg
	return f.err
}

// newTestReconciler wires a reconciler with a fake client seeded with the given
// monitor, plus the provided fake driver/notifier. It returns the reconciler
// and a helper to fetch the persisted monitor.
func newTestReconciler(t *testing.T, monitor *monitorv1alpha1.EndpointMonitor, d driver.Driver, n notifier.Notifier) (*EndpointMonitorReconciler, func() monitorv1alpha1.EndpointMonitor) {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := monitorv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(monitor).
		WithStatusSubresource(monitor).
		Build()

	r := &EndpointMonitorReconciler{
		Client: c,
		Scheme: scheme,
		newDriver: func(driverType, endpoint string, m *monitorv1alpha1.EndpointMonitor) (driver.Driver, error) {
			return d, nil
		},
		newNotifier: func(cfg *monitorv1alpha1.NotifyConfig) (notifier.Notifier, error) {
			return n, nil
		},
	}

	key := types.NamespacedName{Name: monitor.Name, Namespace: monitor.Namespace}
	fetch := func() monitorv1alpha1.EndpointMonitor {
		var got monitorv1alpha1.EndpointMonitor
		if err := c.Get(context.Background(), key, &got); err != nil {
			t.Fatalf("failed to fetch monitor: %v", err)
		}
		return got
	}

	return r, fetch
}

func baseMonitor() *monitorv1alpha1.EndpointMonitor {
	return &monitorv1alpha1.EndpointMonitor{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: monitorv1alpha1.EndpointMonitorSpec{
			Driver:        "fake",
			Endpoint:      "http://fake",
			CheckInterval: 60,
			Notify: monitorv1alpha1.NotifyConfig{
				Slack: &monitorv1alpha1.SlackConfig{Enabled: true, WebhookURL: "http://x"},
			},
		},
	}
}

func request() reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"}}
}

func TestReconcile_SkipsWhenIntervalNotElapsed(t *testing.T) {
	m := baseMonitor()
	m.Status.LastCheckedTime = metav1.NewTime(time.Now())
	m.Status.LastStatus = "success"

	d := &fakeDriver{result: &driver.CheckResult{Success: true}}
	n := &fakeNotifier{}
	r, _ := newTestReconciler(t, m, d, n)

	res, err := r.Reconcile(context.Background(), request())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.called {
		t.Error("driver.Check should NOT be called before checkInterval elapses")
	}
	if n.called {
		t.Error("notifier should NOT be called before checkInterval elapses")
	}
	if res.RequeueAfter <= 0 {
		t.Errorf("expected a positive RequeueAfter, got %v", res.RequeueAfter)
	}
}

func TestReconcile_SuccessUpdatesStatusAndAlerts(t *testing.T) {
	d := &fakeDriver{result: &driver.CheckResult{Success: true, Message: "healthy", ResponseTime: time.Millisecond}}
	n := &fakeNotifier{}
	r, fetch := newTestReconciler(t, baseMonitor(), d, n)

	res, err := r.Reconcile(context.Background(), request())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.called {
		t.Error("expected driver.Check to be called")
	}
	if !n.called || n.gotStatus != "success" {
		t.Errorf("expected notifier called with success, got called=%v status=%q", n.called, n.gotStatus)
	}
	if res.RequeueAfter != 60*time.Second {
		t.Errorf("RequeueAfter = %v, want 60s", res.RequeueAfter)
	}

	got := fetch()
	if got.Status.LastStatus != "success" {
		t.Errorf("LastStatus = %q, want %q", got.Status.LastStatus, "success")
	}
	if got.Status.LastCheckedTime.IsZero() {
		t.Error("expected LastCheckedTime to be set")
	}
}

func TestReconcile_FailureUpdatesStatusAndAlerts(t *testing.T) {
	d := &fakeDriver{result: &driver.CheckResult{Success: false, Message: "unhealthy"}}
	n := &fakeNotifier{}
	r, fetch := newTestReconciler(t, baseMonitor(), d, n)

	if _, err := r.Reconcile(context.Background(), request()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.gotStatus != "failure" {
		t.Errorf("expected notifier called with failure, got %q", n.gotStatus)
	}
	got := fetch()
	if got.Status.LastStatus != "failure" {
		t.Errorf("LastStatus = %q, want %q", got.Status.LastStatus, "failure")
	}
}

// TestReconcile_NotifierFailureDoesNotPersistStatus documents the CURRENT
// behavior: when SendAlert fails, Reconcile returns the error before writing
// status, so the resource is left with no status and the controller requeues
// with backoff (ignoring checkInterval). This test pins that behavior so any
// future change to it is intentional and visible.
func TestReconcile_NotifierFailureDoesNotPersistStatus(t *testing.T) {
	d := &fakeDriver{result: &driver.CheckResult{Success: true, Message: "healthy"}}
	n := &fakeNotifier{err: errors.New("non-200 response from slack: 404 Not Found")}
	r, fetch := newTestReconciler(t, baseMonitor(), d, n)

	_, err := r.Reconcile(context.Background(), request())
	if err == nil {
		t.Fatal("expected Reconcile to return an error when the notifier fails")
	}

	got := fetch()
	if got.Status.LastStatus != "" {
		t.Errorf("expected no status persisted on notifier failure, got %q", got.Status.LastStatus)
	}
	if !got.Status.LastCheckedTime.IsZero() {
		t.Error("expected LastCheckedTime to remain unset on notifier failure")
	}
}
