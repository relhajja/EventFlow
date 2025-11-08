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

	eventflowv1alpha1 "github.com/relhajja/eventflow/operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=eventflow.eventflow.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventflow.eventflow.io,resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eventflow.eventflow.io,resources=functions/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *FunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// 1. Fetch the Function CR
	function := &eventflowv1alpha1.Function{}
	if err := r.Get(ctx, req.NamespacedName, function); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Function resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch Function")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling Function", "name", function.Name, "image", function.Spec.Image)

	// 2. Check if Deployment exists
	deployment := &appsv1.Deployment{}
	deploymentName := fmt.Sprintf("fn-%s", function.Name)
	err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: function.Namespace}, deployment)

	if err != nil && errors.IsNotFound(err) {
		// 3. Create a new Deployment
		deployment, err := r.buildDeployment(function)
		if err != nil {
			logger.Error(err, "Failed to build Deployment for Function", "function", function.Name)
			return ctrl.Result{}, err
		}

		// Set Function as owner of the Deployment (for garbage collection)
		if err := controllerutil.SetControllerReference(function, deployment, r.Scheme); err != nil {
			logger.Error(err, "Failed to set owner reference for Deployment", "function", function.Name)
			return ctrl.Result{}, err
		}

		logger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		if err := r.Create(ctx, deployment); err != nil {
			logger.Error(err, "Failed to create Deployment for Function", "function", function.Name)
			return ctrl.Result{}, err
		}

		// Update Function status
		function.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, function); err != nil {
			logger.Error(err, "Failed to update Function status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil

	} else if err != nil {
		logger.Error(err, "Failed to get Deployment for Function", "function", function.Name)
		return ctrl.Result{}, err
	}

	// 4. Deployment exists, check if update needed
	needsUpdate := false

	// Get desired replicas (default to 1 if not set)
	desiredReplicas := int32(1)
	if function.Spec.Replicas != nil {
		desiredReplicas = *function.Spec.Replicas
	}

	// Check if replicas changed
	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas != desiredReplicas {
		deployment.Spec.Replicas = &desiredReplicas
		needsUpdate = true
	}

	// Check if image changed
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if deployment.Spec.Template.Spec.Containers[0].Image != function.Spec.Image {
			deployment.Spec.Template.Spec.Containers[0].Image = function.Spec.Image
			needsUpdate = true
		}
	}

	if needsUpdate {
		logger.Info("Updating Deployment for Function", "deployment", deployment.Name)
		if err := r.Update(ctx, deployment); err != nil {
			logger.Error(err, "Failed to update Deployment")
			return ctrl.Result{}, err
		}
	}

	// 5. Update Function status with Deployment info
	function.Status.Replicas = deployment.Status.Replicas
	function.Status.AvailableReplicas = deployment.Status.AvailableReplicas

	// Set phase based on replica status
	if deployment.Status.AvailableReplicas == desiredReplicas {
		function.Status.Phase = "Running"
	} else if deployment.Status.AvailableReplicas > 0 {
		function.Status.Phase = "Running" // Partially running is still Running
	} else {
		function.Status.Phase = "Pending"
	}

	// Add condition
	readyStatus := metav1.ConditionFalse
	if deployment.Status.AvailableReplicas > 0 {
		readyStatus = metav1.ConditionTrue
	}

	function.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             readyStatus,
			LastTransitionTime: metav1.Now(),
			Reason:             "DeploymentReady",
			Message:            fmt.Sprintf("%d/%d replicas available", deployment.Status.AvailableReplicas, desiredReplicas),
		},
	}

	if err := r.Status().Update(ctx, function); err != nil {
		logger.Error(err, "Failed to update Function status", "function", function.Name)
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled Function",
		"phase", function.Status.Phase,
		"replicas", fmt.Sprintf("%d/%d", function.Status.AvailableReplicas, function.Spec.Replicas))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventflowv1alpha1.Function{}).
		Owns(&appsv1.Deployment{}). // Watch Deployments owned by Functions
		Named("function").
		Complete(r)
}

// buildDeployment creates a Deployment spec from a Function CR
func (r *FunctionReconciler) buildDeployment(function *eventflowv1alpha1.Function) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app":      "eventflow-function",
		"function": function.Name,
	}

	// Build environment variables
	var envVars []corev1.EnvVar
	for key, value := range function.Spec.Env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// Build container spec
	container := corev1.Container{
		Name:            "function",
		Image:           function.Spec.Image,
		ImagePullPolicy: corev1.PullIfNotPresent, // For kind clusters
		Env:             envVars,
	}

	// Add command if specified
	if len(function.Spec.Command) > 0 {
		container.Command = function.Spec.Command
	}

	// Add args if specified
	if len(function.Spec.Args) > 0 {
		container.Args = function.Spec.Args
	}

	// Add resource requirements (required by tenant resource quotas)
	// Set default values if not specified
	container.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	// Override with custom resources if specified
	if function.Spec.Resources != nil {
		// Note: For simplicity, we're not parsing the resource strings yet
		// In production, you'd parse strings like "100m" and "128Mi" from function.Spec.Resources
	}

	// Get desired replicas (default to 1 if not set)
	replicas := int32(1)
	if function.Spec.Replicas != nil {
		replicas = *function.Spec.Replicas
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("fn-%s", function.Name),
			Namespace: function.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
				},
			},
		},
	}

	return deployment, nil
}
