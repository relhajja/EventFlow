package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ============================================================================
// Domain Models
// ============================================================================

// BuildReq represents a request to build a container image from source
type BuildReq struct {
	BuildID            string            `json:"build_id"`
	SourceType         string            `json:"source_type"` // "git" | "tar" | "code"
	Source             string            `json:"source"`      // Git URL, tar URL, or inline code
	ImageRef           string            `json:"image_ref"`   // Target image name with registry
	Prefer             *string           `json:"prefer,omitempty"`
	GitRef             string            `json:"git_ref,omitempty"`    // Branch/tag for git sources
	Runtime            string            `json:"runtime,omitempty"`    // python, node, go, java, etc.
	Env                map[string]string `json:"env,omitempty"`        // Environment variables
	RegistrySecretName string            `json:"registry_secret_name"` // Kubernetes secret for registry auth
	TimeoutSeconds     int32             `json:"timeout_seconds,omitempty"`
}

// Status represents the current state of a build
type Status struct {
	BuildID  string `json:"build_id"`
	Event    string `json:"event"` // started, building, complete, failed
	Message  string `json:"message,omitempty"`
	Strategy string `json:"strategy,omitempty"` // cnb (Cloud Native Buildpacks)
	ImageRef string `json:"image_ref,omitempty"`
	Digest   string `json:"digest,omitempty"` // Image SHA256 digest
}

// ============================================================================
// Configuration Constants
// ============================================================================

const (
	// Environment variable names for configuration
	nsEnv             = "NAMESPACE"       // Kubernetes namespace
	builderSAEnv      = "BUILDER_SA"      // Service account for build Jobs
	kanikoImageEnv    = "KANIKO_IMAGE"    // Kaniko executor image (unused in CNB mode)
	packImageEnv      = "PACK_IMAGE"      // Pack CLI image (unused, using docker:cli)
	builderImageEnv   = "BUILDER_IMAGE"   // CNB builder image
	registrySecretEnv = "REGISTRY_SECRET" // Registry credentials secret

	// Build strategy
	strategyCloudNativeBuildpacks = "cnb"

	// Default values
	defaultNATSURL        = "nats://nats.eventflow.svc.cluster.local:4222"
	defaultBuilderSA      = "builder"
	defaultRegistrySecret = "registry-secret"
	defaultCNBBuilder     = "paketobuildpacks/builder-jammy-base:latest"

	// Job configuration
	jobTTLSeconds       = 600 // Clean up completed jobs after 10 minutes
	jobBackoffLimit     = 0   // Don't retry failed builds
	buildTimeout        = 10 * time.Minute
	statusCheckInterval = 5 * time.Second
)

// ============================================================================
// Main Application Entry Point
// ============================================================================

func main() {
	namespace := mustEnv(nsEnv)
	log.Printf("Builder worker starting in namespace: %s", namespace)

	// Connect to NATS messaging system
	natsURL := getEnvOrDefault("NATS_URL", defaultNATSURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS at %s: %v", natsURL, err)
	}
	defer nc.Close()

	// Subscribe to build requests
	_, err = nc.Subscribe("eventflow.events", func(msg *nats.Msg) {
		handleBuildRequest(nc, msg.Data)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to events: %v", err)
	}

	log.Println("Worker listening on eventflow.events")
	select {} // Block forever
}

// handleBuildRequest processes incoming build request messages
func handleBuildRequest(nc *nats.Conn, data []byte) {
	var req BuildReq
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("Failed to unmarshal build request: %v", err)
		return
	}

	log.Printf("Received build request for %s (source: %s, runtime: %s)",
		req.BuildID, req.SourceType, req.Runtime)

	if err := processBuild(nc, req); err != nil {
		log.Printf("Build failed for %s: %v", req.BuildID, err)
		publishStatus(nc, req.BuildID, "failed", err.Error(), "", "", "")
	}
}

// ============================================================================
// Build Processing
// ============================================================================

// processBuild handles the complete build lifecycle
func processBuild(nc *nats.Conn, req BuildReq) error {
	ctx := context.Background()
	clientset, namespace := getKubernetesClient()

	// Determine build strategy (always CNB for now)
	strategy := strategyCloudNativeBuildpacks
	log.Printf("Using build strategy: %s", strategy)
	publishStatus(nc, req.BuildID, "started",
		fmt.Sprintf("Starting build with %s", strategy), strategy, "", "")

	// Create Kubernetes Job for the build
	job := createBuildJob(namespace, req, strategy)
	created, err := clientset.BatchV1().Jobs(namespace).Create(ctx, job, meta.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create build job: %w", err)
	}

	log.Printf("Created build job: %s", created.Name)
	publishStatus(nc, req.BuildID, "building",
		"Build job created, building image...", strategy, req.ImageRef, "")

	// Wait for build completion
	if err := waitForJobCompletion(ctx, clientset, namespace, created.Name, buildTimeout); err != nil {
		return fmt.Errorf("build job failed: %w", err)
	}

	publishStatus(nc, req.BuildID, "complete", "Build succeeded",
		strategy, req.ImageRef, "sha256:placeholder")
	return nil
}

// ============================================================================
// Kubernetes Job Creation
// ============================================================================

// createBuildJob creates a Kubernetes Job spec for building a container image
func createBuildJob(namespace string, req BuildReq, strategy string) *batchv1.Job {
	ttl := int32(jobTTLSeconds)
	backoff := int32(jobBackoffLimit)
	jobName := fmt.Sprintf("build-%s", req.BuildID[:8])

	// Get registry secret name (for pushing images)
	registrySecret := req.RegistrySecretName
	if registrySecret == "" {
		registrySecret = getEnvOrDefault(registrySecretEnv, defaultRegistrySecret)
	}

	// Get service account for the Job pods
	builderSA := getEnvOrDefault(builderSAEnv, defaultBuilderSA)

	job := &batchv1.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":      "builder",
				"build-id": req.BuildID,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoff,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: builderSA,
					RestartPolicy:      corev1.RestartPolicyNever,
					InitContainers:     []corev1.Container{},
					Containers:         []corev1.Container{},
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "docker-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: registrySecret,
									Items: []corev1.KeyToPath{
										{Key: ".dockerconfigjson", Path: "config.json"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add fetch init container
	job.Spec.Template.Spec.InitContainers = append(
		job.Spec.Template.Spec.InitContainers,
		fetchContainer(req),
	)

	// Always use Cloud Native Buildpacks with DinD
	// Add DinD sidecar
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		dindContainer(),
	)
	// Add Pack container
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		cnbContainer(req),
	)

	return job
}

// ============================================================================
// Container Definitions (Init and Main Containers)
// ============================================================================

// fetchContainer creates an init container that fetches source code
// Supports three source types: git (clone repo), tar (download/extract), code (inline)
func fetchContainer(req BuildReq) corev1.Container {
	container := corev1.Container{
		Name:  "fetch",
		Image: "alpine/git:latest",
		VolumeMounts: []corev1.VolumeMount{
			{Name: "workspace", MountPath: "/workspace"},
		},
	}

	switch req.SourceType {
	case "git":
		// Clone a Git repository
		gitRef := req.GitRef
		if gitRef == "" {
			gitRef = "main"
		}
		container.Command = []string{"sh", "-c"}
		container.Args = []string{
			fmt.Sprintf("git clone --depth 1 --branch %s %s /workspace", gitRef, req.Source),
		}

	case "tar":
		// Download and extract a tarball
		container.Image = "busybox:latest"
		container.Command = []string{"sh", "-c"}
		container.Args = []string{
			fmt.Sprintf("wget -O /tmp/source.tar.gz %s && tar -xzf /tmp/source.tar.gz -C /workspace", req.Source),
		}

	case "code":
		// Write inline code to a file (for code editor deployments)
		container.Image = "busybox:latest"
		container.Command = []string{"sh", "-c"}
		container.Args = []string{
			fmt.Sprintf("mkdir -p /workspace && cat > /workspace/main.py << 'EOFCODE'\n%s\nEOFCODE", req.Source),
		}
	}

	return container
}

// cnbContainer creates the main container that runs Cloud Native Buildpacks
// Uses docker:cli image with pack CLI downloaded at runtime
func cnbContainer(req BuildReq) corev1.Container {
	// Select appropriate CNB builder based on runtime
	builderImage := getEnvOrDefault(builderImageEnv, defaultCNBBuilder)

	if req.Runtime != "" {
		switch req.Runtime {
		case "python", "python3":
			builderImage = "paketobuildpacks/builder-jammy-base:latest"
		case "node", "nodejs":
			builderImage = "paketobuildpacks/builder-jammy-base:latest"
		case "go":
			builderImage = "paketobuildpacks/builder-jammy-tiny:latest"
		case "java":
			builderImage = "paketobuildpacks/builder-jammy-base:latest"
		}
	}

	return corev1.Container{
		Name:    "pack",
		Image:   "docker:27-cli",
		Command: []string{"/bin/sh", "-c"},
		Args: []string{
			`
			set -e
			
			echo "Waiting for Docker daemon..."
			until docker -H tcp://localhost:2375 info > /dev/null 2>&1; do
				sleep 2
			done
			echo "Docker daemon is ready!"
			
			# Download pack CLI
			echo "Downloading pack CLI..."
			wget -q -O - https://github.com/buildpacks/pack/releases/download/v0.35.1/pack-v0.35.1-linux.tgz | tar -xz -C /usr/local/bin
			chmod +x /usr/local/bin/pack
			
			# Run pack build
			echo "Running pack build..."
			pack build ` + req.ImageRef + ` \
				--path /workspace \
				--builder ` + builderImage + ` \
				--publish \
				--docker-host tcp://localhost:2375 \
				--trust-builder
			
			echo "Build complete!"
			`,
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "workspace", MountPath: "/workspace"},
		},
		Env: []corev1.EnvVar{
			{Name: "DOCKER_HOST", Value: "tcp://localhost:2375"},
		},
	}
}

// dindContainer creates a Docker-in-Docker sidecar container
// Provides a Docker daemon that the pack container can use for building images
func dindContainer() corev1.Container {
	return corev1.Container{
		Name:  "dind",
		Image: "docker:24-dind",
		Env: []corev1.EnvVar{
			{Name: "DOCKER_TLS_CERTDIR", Value: ""}, // Disable TLS for local communication
		},
		SecurityContext: &corev1.SecurityContext{
			Privileged: boolPtr(true), // Required for DinD to manage containers
		},
		Args: []string{
			"--insecure-registry=docker-registry.eventflow.svc.cluster.local:5000",
		},
	}
}

// ============================================================================
// Job Monitoring
// ============================================================================

// waitForJobCompletion waits for a Kubernetes Job to complete or fail
func waitForJobCompletion(ctx context.Context, clientset *kubernetes.Clientset, namespace, jobName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(statusCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for job to complete")
		case <-ticker.C:
			job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, meta.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get job status: %w", err)
			}

			if job.Status.Succeeded > 0 {
				log.Printf("Job %s completed successfully", jobName)
				return nil
			}

			if job.Status.Failed > 0 {
				return fmt.Errorf("job %s failed", jobName)
			}
		}
	}
}

// ============================================================================
// NATS Event Publishing
// ============================================================================

// publishStatus sends a build status update to NATS
func publishStatus(nc *nats.Conn, buildID, event, message, strategy, imageRef, digest string) {
	status := Status{
		BuildID:  buildID,
		Event:    event,
		Message:  message,
		Strategy: strategy,
		ImageRef: imageRef,
		Digest:   digest,
	}

	data, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal status: %v", err)
		return
	}

	subject := fmt.Sprintf("builds.status.%s", buildID)
	if err := nc.Publish(subject, data); err != nil {
		log.Printf("Failed to publish status to %s: %v", subject, err)
	}
}

// ============================================================================
// Kubernetes Client Management
// ============================================================================

// getKubernetesClient creates a Kubernetes client and returns it with the namespace
func getKubernetesClient() (*kubernetes.Clientset, string) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig for local development
		kubeconfigPath := os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			log.Fatalf("Failed to get Kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	namespace := mustEnv(nsEnv)
	return clientset, namespace
}

// ============================================================================
// Helper Functions
// ============================================================================

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// mustEnv returns the value of an environment variable or panics if not set
func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable not set: %s", key)
	}
	return val
}

// boolPtr returns a pointer to a boolean value (helper for Kubernetes API)
func boolPtr(b bool) *bool {
	return &b
}
