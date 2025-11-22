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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/example/nginx-operator/api/v1alpha1"
	"github.com/example/nginx-operator/assets"
)

// NginxOperatorReconciler reconciles a NginxOperator object
type NginxOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var deploymentManifest = assets.
	GetDeploymentFromFile("manifests/nginx_deployment.yaml")

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *NginxOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	operatorCR := &operatorv1alpha1.NginxOperator{}
	err := r.Get(ctx, req.NamespacedName, operatorCR)
	if err != nil {
		if errors.IsNotFound(err) {
			// The CR was deleted before we processed it. Nothing to do.
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error getting operator resource object")
		return ctrl.Result{}, err
	}

	deploymentManifest := assets.GetDeploymentFromFile("manifests/nginx_deployment.yaml")
	deploymentManifest.Namespace = operatorCR.Namespace

	if operatorCR.Spec.Replicas != nil {
		deploymentManifest.Spec.Replicas = operatorCR.Spec.Replicas
	}

	if operatorCR.Spec.Port != nil &&
		len(deploymentManifest.Spec.Template.Spec.Containers) > 0 &&
		len(deploymentManifest.Spec.Template.Spec.Containers[0].Ports) > 0 {
		deploymentManifest.Spec.Template.Spec.Containers[0].
			Ports[0].ContainerPort = *operatorCR.Spec.Port
	}

	if err := ctrl.SetControllerReference(operatorCR, deploymentManifest, r.Scheme); err != nil {
		logger.Error(err, "Error setting owner reference on deployment")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, deploymentManifest); err != nil {
		if errors.IsAlreadyExists(err) {
			existingDeployment := &appsv1.Deployment{}
			if getErr := r.Get(
				ctx,
				types.NamespacedName{
					Namespace: deploymentManifest.Namespace,
					Name:      deploymentManifest.Name,
				},
				existingDeployment,
			); getErr != nil {
				logger.Error(getErr, "Error fetching existing Nginx deployment")
				return ctrl.Result{}, getErr
			}

			existingDeployment.Spec = deploymentManifest.Spec
			if updateErr := r.Update(ctx, existingDeployment); updateErr != nil {
				logger.Error(updateErr, "Error updating Nginx deployment.")
				return ctrl.Result{}, updateErr
			}
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error creating Nginx deployment.")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NginxOperator{}).
		Named("nginxoperator").
		Owns(deploymentManifest).
		Complete(r)
}
