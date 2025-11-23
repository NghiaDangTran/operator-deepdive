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
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/example/nginx-operator/api/v1alpha1"
	"github.com/example/nginx-operator/assets"
	"github.com/example/nginx-operator/internal/controller/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
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
	metrics.ReconcilesTotal.Inc()
	logger := logf.FromContext(ctx)
	operatorCR := &operatorv1alpha1.NginxOperator{}
	// Fetch the NginxOperator custom resource to drive desired state.
	err := r.Get(ctx, req.NamespacedName, operatorCR)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error getting operator resource object")
		meta.SetStatusCondition(&operatorCR.Status.Conditions, metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionTrue,
			Reason:             "OperatorResourceNotAvailable",
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            fmt.Sprintf("unable to getoperator custom resource: %s", err.Error()),
		})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
	}

	// Load the desired Deployment template from assets and scope it to the CR namespace.
	deploymentManifest := assets.GetDeploymentFromFile("manifests/nginx_deployment.yaml")
	deploymentManifest.Namespace = operatorCR.Namespace

	// Override replicas from the CR spec if provided.
	if operatorCR.Spec.Replicas != nil {
		deploymentManifest.Spec.Replicas = operatorCR.Spec.Replicas
	}

	// Override container port from the CR spec if provided.
	if operatorCR.Spec.Port != nil &&
		len(deploymentManifest.Spec.Template.Spec.Containers) > 0 &&
		len(deploymentManifest.Spec.Template.Spec.Containers[0].Ports) > 0 {
		deploymentManifest.Spec.Template.Spec.Containers[0].
			Ports[0].ContainerPort = *operatorCR.Spec.Port
	}

	// Set owner reference so the Deployment is garbage-collected with the CR.
	if err := ctrl.SetControllerReference(operatorCR, deploymentManifest, r.Scheme); err != nil {
		logger.Error(err, "Error setting owner reference on deployment")
		return ctrl.Result{}, err
	}

	// Create Deployment if missing; otherwise update it with conflict retry.
	if err := r.Create(ctx, deploymentManifest); err != nil {
		if errors.IsAlreadyExists(err) {
			// Deployment exists; fetch latest and retry updates to avoid resourceVersion conflicts.
			existingDeployment := &appsv1.Deployment{}
			namespacedName := types.NamespacedName{
				Namespace: deploymentManifest.Namespace,
				Name:      deploymentManifest.Name,
			}
			// pull the current Deployment (existingDeployment) so we know it exists; if this fetch fails we bail out.
			if getErr := r.Get(ctx, namespacedName, existingDeployment); getErr != nil {
				logger.Error(getErr, "Error fetching existing Nginx deployment")
				return ctrl.Result{}, getErr
			}
			//   this is the spec we want applied.
			desiredSpec := deploymentManifest.Spec
			updateErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				current := &appsv1.Deployment{}
				if err := r.Get(ctx, namespacedName, current); err != nil {
					return err
				}

				current.Spec = desiredSpec
				return r.Update(ctx, current)
			})
			if updateErr != nil {
				logger.Error(updateErr, "Error updating Nginx deployment after conflict retry.")
				meta.SetStatusCondition(&operatorCR.Status.Conditions,
					metav1.Condition{
						Type:               "OperatorDegraded",
						Status:             metav1.ConditionTrue,
						Reason:             "OperandUpdateDeploymentFailed",
						LastTransitionTime: metav1.NewTime(time.Now()),
						Message:            fmt.Sprintf("unable to update deployment: %s", updateErr.Error()),
					})
				return ctrl.Result{}, utilerrors.NewAggregate([]error{updateErr, r.Status().Update(ctx, operatorCR)})
			}
			return ctrl.Result{}, nil
		}
		meta.SetStatusCondition(&operatorCR.Status.Conditions,
			metav1.Condition{
				Type:               "OperatorDegraded",
				Status:             metav1.ConditionTrue,
				Reason:             "OperandDeploymentNotAvailable",
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            fmt.Sprintf("unable to get operand deployment: %s", err.Error()),
			})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, operatorCR)})
	}

	meta.SetStatusCondition(&operatorCR.Status.Conditions,
		metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionFalse,
			Reason:             "OperatorSucceeded",
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            "operator successfully reconciling",
		})
	return ctrl.Result{}, utilerrors.NewAggregate([]error{err,
		r.Status().Update(ctx, operatorCR)})
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NginxOperator{}).
		Named("nginxoperator").
		Owns(deploymentManifest).
		Complete(r)
}
