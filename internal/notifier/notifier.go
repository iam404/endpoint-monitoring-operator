package notifier

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Notifier interface {
	SendAlert(status string, values *NoticeValues, client client.Client) error
}

type NoticeValues struct {
	Status           string
	Endpoint         string
	Driver           string
	Healthy          string
	Name             string
	Response         string
	LastStatus       string
	LastChecked      string
	LastStatusChange string
	CurrentTime      string
	StatusTime       string
	Message          string
	Namespace        string
	AlertMessage     string
}

// ShouldAlert returns true if the given status is in the allowed list, or if the list is empty and status is "failure".
func ShouldAlert(allowed []string, status string) bool {
	if len(allowed) == 0 {
		// Default to alert only on failure
		return status == "failure"
	}
	for _, a := range allowed {
		if a == status {
			return true
		}
	}
	return false
}

func GetSecret(name string, namespace string, client client.Client) (*corev1.Secret, error) {
	secretName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), secretName, secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}
