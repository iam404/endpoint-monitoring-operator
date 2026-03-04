package scheduler

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	monitoringv1alpha1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/driver"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
	"k8s.io/apimachinery/pkg/types"
	clocktesting "k8s.io/utils/clock/testing"
)

type countingDriver struct {
	key    string
	counts *sync.Map
}

func (d *countingDriver) Check() (*driver.CheckResult, error) {
	v, _ := d.counts.LoadOrStore(d.key, int64(0))
	d.counts.Store(d.key, v.(int64)+1)
	return &driver.CheckResult{Success: true, ResponseTime: 10 * time.Millisecond, StatusCode: 200, Message: "ok"}, nil
}
func (d *countingDriver) GetEndpoint() string { return d.key }
func (d *countingDriver) GetType() string     { return "mock" }

type fakeDriverFactory struct{ counts *sync.Map }

func (f *fakeDriverFactory) CreateDriver(_ string, endpoint string, _ *monitoringv1alpha1.EndpointMonitor) (driver.Driver, error) {
	return &countingDriver{key: endpoint, counts: f.counts}, nil
}

type fakeNotifier struct{}

func (f *fakeNotifier) SendAlert(string, string) error { return nil }

type fakeNotifierFactory struct{}

func (f *fakeNotifierFactory) CreateNotifier(_ *monitoringv1alpha1.NotifyConfig) (notifier.Notifier, error) {
	return &fakeNotifier{}, nil
}

type fakeStatusWriter struct{}

func (f *fakeStatusWriter) Write(context.Context, types.NamespacedName, ProbeResult) error {
	return nil
}

func TestSchedulerRunsThreeProbesOnFakeClock(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	counts := &sync.Map{}

	s := newWithClock(
		Config{Workers: 4, QueueSize: 100},
		&fakeDriverFactory{counts: counts},
		&fakeNotifierFactory{},
		&fakeStatusWriter{},
		fakeClock,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = s.Start(ctx) }()

	interval := 2
	for i := 0; i < 3; i++ {
		key := types.NamespacedName{Name: "probe", Namespace: string(rune('a' + i))}
		s.Upsert(key, monitoringv1alpha1.EndpointMonitorSpec{
			Driver:        "http",
			Endpoint:      key.Namespace,
			CheckInterval: interval,
			Notify:        monitoringv1alpha1.NotifyConfig{Slack: &monitoringv1alpha1.SlackConfig{Enabled: true, WebhookURL: "http://example"}},
		})
	}

	fakeClock.Step(2 * time.Second)
	fakeClock.Step(2 * time.Second)

	waitFor(t, 2*time.Second, func() bool {
		for _, endpoint := range []string{"a", "b", "c"} {
			v, ok := counts.Load(endpoint)
			if !ok || v.(int64) < 2 {
				return false
			}
		}
		return true
	})
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		runtime.Gosched()
	}
	t.Fatalf("condition was not met before timeout %s", timeout)
}
