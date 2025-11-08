package k8s

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/eventflow/api/internal/models"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

// NewClient creates a new Kubernetes client
// Returns a client in demo mode if Kubernetes is not available
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

		// Check if kubeconfig exists and is a file
		if info, statErr := os.Stat(kubeconfig); statErr == nil && !info.IsDir() {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				log.Printf("Warning: Failed to load kubeconfig: %v. Running in demo mode.", err)
				return &Client{
					clientset: nil, // Demo mode - no real K8s
				}, nil
			}
		} else {
			log.Printf("Warning: No valid kubeconfig found. Running in demo mode.")
			return &Client{
				clientset: nil, // Demo mode - no real K8s
			}, nil
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create dynamic client for CRDs
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	log.Printf("✅ Kubernetes client initialized")

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}, nil
}

// CreateDeployment creates a new deployment for a function
func (c *Client) CreateDeployment(ctx context.Context, namespace, name, image string, replicas int32, env map[string]string, command []string) error {
	if c.clientset == nil {
		log.Printf("Demo mode: Creating function %s with image %s", name, image)
		return nil
	}

	envVars := []corev1.EnvVar{}
	for k, v := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	labels := map[string]string{
		"app":        "eventflow-function",
		"function":   name,
		"managed-by": "eventflow",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("fn-%s", name),
			Namespace: namespace,
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
					Containers: []corev1.Container{
						{
							Name:    name,
							Image:   image,
							Command: command,
							Env:     envVars,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "http",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Update existing deployment
			_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		}
		return err
	}

	// Create service
	return c.createService(ctx, namespace, name, labels)
}

// createService creates a service for the function
func (c *Client) createService(ctx context.Context, namespace, name string, labels map[string]string) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("fn-%s", name),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// GetDeployment retrieves deployment details
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("demo mode: deployment %s not found", name)
	}
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, fmt.Sprintf("fn-%s", name), metav1.GetOptions{})
}

// ListDeployments lists all function deployments
func (c *Client) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	if c.clientset == nil {
		// Demo mode - return empty list
		return &appsv1.DeploymentList{
			Items: []appsv1.Deployment{},
		}, nil
	}
	return c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=eventflow-function",
	})
}

// DeleteDeployment deletes a function deployment and service
func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string) error {
	if c.clientset == nil {
		log.Printf("Demo mode: Would delete deployment for function %s", name)
		return nil
	}

	deploymentName := fmt.Sprintf("fn-%s", name)

	// Delete deployment
	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Delete service
	err = c.clientset.CoreV1().Services(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// CreateJob creates a one-time job for function invocation
func (c *Client) CreateJob(ctx context.Context, namespace, name, image string, command []string, env map[string]string) error {
	if c.clientset == nil {
		log.Printf("Demo mode: Would create job for function %s", name)
		return nil
	}

	envVars := []corev1.EnvVar{}
	for k, v := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("fn-%s-invoke-", name),
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "invoke",
							Image:   image,
							Command: command,
							Env:     envVars,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: new(int32),
		},
	}
	*job.Spec.BackoffLimit = 0

	_, err := c.clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

// GetPodLogs retrieves logs from pods of a function
func (c *Client) GetPodLogs(ctx context.Context, namespace, name string, follow bool) (io.ReadCloser, error) {
	if c.clientset == nil {
		return io.NopCloser(nil), fmt.Errorf("demo mode: no logs available for function %s", name)
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("function=%s", name),
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for function %s", name)
	}

	// Get logs from the first pod
	pod := pods.Items[0]
	opts := &corev1.PodLogOptions{
		Follow: follow,
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, opts)
	return req.Stream(ctx)
}

// HasKubernetes returns true if running with real Kubernetes
func (c *Client) HasKubernetes() bool {
	return c.clientset != nil
}

// EnsureNamespace creates a namespace if it doesn't exist with resource quotas
func (c *Client) EnsureNamespace(ctx context.Context, namespace string) error {
	if c.clientset == nil {
		log.Printf("Demo mode: Would create namespace %s", namespace)
		return nil
	}

	// Check if namespace exists
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		// Namespace already exists
		log.Printf("Namespace %s already exists", namespace)
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check namespace: %w", err)
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"managed-by": "eventflow",
				"type":       "tenant",
			},
		},
	}

	_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	log.Printf("✅ Created tenant namespace: %s", namespace)

	// Create resource quota for the namespace
	err = c.createResourceQuota(ctx, namespace)
	if err != nil {
		log.Printf("Warning: Failed to create resource quota for %s: %v", namespace, err)
	}

	return nil
}

// createResourceQuota creates resource limits for a tenant namespace
func (c *Client) createResourceQuota(ctx context.Context, namespace string) error {
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourcePods:                   resource.MustParse("50"),
				corev1.ResourceRequestsCPU:            resource.MustParse("10"),
				corev1.ResourceRequestsMemory:         resource.MustParse("20Gi"),
				corev1.ResourceLimitsCPU:              resource.MustParse("20"),
				corev1.ResourceLimitsMemory:           resource.MustParse("40Gi"),
				corev1.ResourcePersistentVolumeClaims: resource.MustParse("10"),
			},
		},
	}

	_, err := c.clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource quota: %w", err)
	}

	log.Printf("✅ Created resource quota for namespace: %s", namespace)
	return nil
}

// InvokeFunction makes an HTTP request to the function's service
func (c *Client) InvokeFunction(ctx context.Context, namespace, name string, method string, path string, body io.Reader) (*corev1.Pod, []byte, error) {
	if c.clientset == nil {
		return nil, []byte(`{"message":"demo mode - function invoked","name":"` + name + `"}`), nil
	}

	// Get pods for the function
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("function=%s", name),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, nil, fmt.Errorf("no pods found for function %s", name)
	}

	// Find a ready pod
	var readyPod *corev1.Pod
	for i := range pods.Items {
		pod := &pods.Items[i]
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				readyPod = pod
				break
			}
		}
		if readyPod != nil {
			break
		}
	}

	if readyPod == nil {
		return nil, nil, fmt.Errorf("no ready pods found for function %s", name)
	}

	return readyPod, []byte(`{"message":"function invoked via pod","pod":"` + readyPod.Name + `"}`), nil
}

func (c *Client) GetDynamicClient(ctx context.Context, req models.CreateFunctionRequest) (dynamic.Interface, error) {
	if c.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client is not initialized")
	}
	return c.dynamicClient, nil
}

func (c *Client) CreateFunctionCR(ctx context.Context, req models.CreateFunctionRequest) error {
	if c.dynamicClient == nil {
		return fmt.Errorf("dynamic client is not initialized")
	}

	functionGVR := schema.GroupVersionResource{
		Group:    "eventflow.eventflow.io",
		Version:  "v1alpha1",
		Resource: "functions",
	}
	spec := map[string]interface{}{
		"image":    req.Image,
		"replicas": req.Replicas,
	}
	if len(req.Command) > 0 {
		spec["command"] = req.Command
	}
	if len(req.Env) > 0 {
		spec["env"] = req.Env
	}

	function := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": functionGVR.GroupVersion().String(),
			"kind":       "Function",
			"metadata": map[string]interface{}{
				"name":      req.Name,
				"namespace": req.Namespace,
			},
			"labels": map[string]interface{}{
				"app":        "eventflow",
				"managed-by": "eventflow-api",
			},
			"spec": spec,
		},
	}

	_, err := c.dynamicClient.Resource(functionGVR).Namespace(req.Namespace).Create(ctx, function, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Function CR: %w", err)
	}

	return nil
}
