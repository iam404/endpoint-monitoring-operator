package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitorv1alpha1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/scheduler"
)

const probeCleanupFinalizer = "monitoring.licious.app/probe-cleanup"

type EndpointMonitorReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler scheduler.ProbeScheduler
}

func (r *EndpointMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var monitor monitorv1alpha1.EndpointMonitor
	if err := r.Get(ctx, req.NamespacedName, &monitor); err != nil {
		if apierrors.IsNotFound(err) {
			r.Scheduler.Delete(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if monitor.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&monitor, probeCleanupFinalizer) {
			controllerutil.AddFinalizer(&monitor, probeCleanupFinalizer)
			if err := r.Update(ctx, &monitor); err != nil {
				return ctrl.Result{}, err
			}
		}
		r.Scheduler.Upsert(req.NamespacedName, monitor.Spec)
		return ctrl.Result{}, nil
	}

	r.Scheduler.Delete(req.NamespacedName)
	if controllerutil.ContainsFinalizer(&monitor, probeCleanupFinalizer) {
		controllerutil.RemoveFinalizer(&monitor, probeCleanupFinalizer)
		if err := r.Update(ctx, &monitor); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *EndpointMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorv1alpha1.EndpointMonitor{}).
		Complete(r)
}
