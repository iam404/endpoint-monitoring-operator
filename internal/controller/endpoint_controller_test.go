/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	Endpointmonitoringv1alpha1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
)

type fakeScheduler struct{}

func (f *fakeScheduler) Upsert(types.NamespacedName, Endpointmonitoringv1alpha1.EndpointMonitorSpec) {
}
func (f *fakeScheduler) Delete(types.NamespacedName) {}
func (f *fakeScheduler) Start(context.Context) error { return nil }

var _ = Describe("EndpointMonitor Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		Endpointmonitor := &Endpointmonitoringv1alpha1.EndpointMonitor{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind EndpointMonitor")
			err := k8sClient.Get(ctx, typeNamespacedName, Endpointmonitor)
			if err != nil && errors.IsNotFound(err) {
				resource := &Endpointmonitoringv1alpha1.EndpointMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: Endpointmonitoringv1alpha1.EndpointMonitorSpec{
						Driver:        "http",
						Endpoint:      "https://example.com",
						CheckInterval: 30,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &Endpointmonitoringv1alpha1.EndpointMonitor{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance EndpointMonitor")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &EndpointMonitorReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				Scheduler: &fakeScheduler{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
