package scheduler

import (
	"context"
	"encoding/json"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1alpha1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
)

// EndpointMonitorStatusWriter patches EndpointMonitor status with probe results.
type EndpointMonitorStatusWriter struct {
	client client.Client
}

// NewEndpointMonitorStatusWriter creates a CR status writer implementation.
func NewEndpointMonitorStatusWriter(c client.Client) *EndpointMonitorStatusWriter {
	return &EndpointMonitorStatusWriter{client: c}
}

// Write patches the EndpointMonitor status for the given key.
func (w *EndpointMonitorStatusWriter) Write(ctx context.Context, key types.NamespacedName, result ProbeResult) error {
	status := map[string]any{
		"lastCheckedTime":   result.CheckedAt,
		"lastStatus":        "failure",
		"lastLatencyMillis": result.Latency.Milliseconds(),
		"lastStatusCode":    result.StatusCode,
		"lastErrorMessage":  result.ErrorMessage,
	}
	if result.Success {
		status["lastStatus"] = "success"
	}

	payload := map[string]any{"status": status}
	patch, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	obj := &monitoringv1alpha1.EndpointMonitor{}
	obj.Name = key.Name
	obj.Namespace = key.Namespace
	if err := w.client.Status().Patch(ctx, obj, client.RawPatch(types.MergePatchType, patch)); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}
