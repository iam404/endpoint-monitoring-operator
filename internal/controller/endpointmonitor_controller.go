package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitorv1alpha1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/pkg/factory"
)

type EndpointMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EndpointMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconcile triggered",
		"name", req.Name,
		"namespace", req.Namespace,
		"timestamp", time.Now().Format(time.RFC3339))

	var monitor monitorv1alpha1.EndpointMonitor
	if err := r.Get(ctx, req.NamespacedName, &monitor); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("EndpointMonitor resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get EndpointMonitor")
		return ctrl.Result{}, err
	}

	now := time.Now()
	checkInterval := time.Duration(monitor.Spec.CheckInterval) * time.Second
	nextCheckTime := monitor.Status.LastCheckedTime.Time.Add(checkInterval)

	if !monitor.Status.LastCheckedTime.IsZero() && now.Before(nextCheckTime) {
		requeueAfter := time.Until(nextCheckTime)
		logger.Info("Skipping reconcile; checkInterval not yet elapsed",
			"name", monitor.Name,
			"lastChecked", monitor.Status.LastCheckedTime.Time,
			"nextCheckDue", nextCheckTime,
			"requeueAfter", requeueAfter)
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	driver, err := factory.NewDriver(monitor.Spec.Driver, monitor.Spec.Endpoint, &monitor)
	if err != nil {
		logger.Error(err, "Failed to create driver")
		return ctrl.Result{}, err
	}

	notifier, err := factory.NewNotifier(&monitor.Spec.Notify)
	if err != nil {
		logger.Error(err, "Failed to create notifier")
		return ctrl.Result{}, err
	}

	result, err := driver.Check()
	if err != nil {
		logger.Error(err, "Failed to perform health check")
		return ctrl.Result{}, err
	}

	var alertMessage, status string
	if result.Success {
		status = "success"
		alertMessage = fmt.Sprintf("%s monitor for %s is healthy\n%s",
			driver.GetType(), driver.GetEndpoint(), result.Message)
	} else {
		status = "failure"
		alertMessage = fmt.Sprintf("%s monitor for %s is unhealthy\n%s",
			driver.GetType(), driver.GetEndpoint(), result.Message)
	}

	if err := notifier.SendAlert(status, alertMessage); err != nil {
		logger.Error(err, "Failed to send alert")
		return ctrl.Result{}, err
	}

	updated := false
	nowMetaTime := metav1.NewTime(now)
	if monitor.Status.LastStatus != status {
		monitor.Status.LastStatus = status
		updated = true
	}
	if monitor.Status.LastCheckedTime != nowMetaTime {
		monitor.Status.LastCheckedTime = nowMetaTime
		updated = true
	}

	if updated {
		if err := r.Status().Update(ctx, &monitor); err != nil {
			logger.Error(err, "Failed to update EndpointMonitor status")
			return ctrl.Result{}, err
		}
	}

	logger.Info("Reconciliation complete",
		"name", monitor.Name,
		"status", status,
		"responseTime", result.ResponseTime.String())

	return ctrl.Result{RequeueAfter: checkInterval}, nil
}

func (r *EndpointMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorv1alpha1.EndpointMonitor{}).
		Complete(r)
}
