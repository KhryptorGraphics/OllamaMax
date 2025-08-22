package deployment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesManager manages Kubernetes deployments
type KubernetesManager struct {
	client    kubernetes.Interface
	logger    *slog.Logger
	config    *K8sConfig
	namespace string
}

// K8sConfig holds Kubernetes deployment configuration
type K8sConfig struct {
	Kubeconfig   string            `yaml:"kubeconfig" json:"kubeconfig"`
	Namespace    string            `yaml:"namespace" json:"namespace"`
	Image        string            `yaml:"image" json:"image"`
	Replicas     int32             `yaml:"replicas" json:"replicas"`
	Resources    K8sResourceLimits `yaml:"resources" json:"resources"`
	Environment  map[string]string `yaml:"environment" json:"environment"`
	Labels       map[string]string `yaml:"labels" json:"labels"`
	ServiceType  string            `yaml:"service_type" json:"service_type"`
	Ports        []K8sPort         `yaml:"ports" json:"ports"`
	HealthCheck  K8sHealthCheck    `yaml:"health_check" json:"health_check"`
	Storage      []K8sVolume       `yaml:"storage" json:"storage"`
}

// K8sResourceLimits defines Kubernetes resource limits
type K8sResourceLimits struct {
	CPURequest    string `yaml:"cpu_request" json:"cpu_request"`
	CPULimit      string `yaml:"cpu_limit" json:"cpu_limit"`
	MemoryRequest string `yaml:"memory_request" json:"memory_request"`
	MemoryLimit   string `yaml:"memory_limit" json:"memory_limit"`
}

// K8sPort defines a port configuration
type K8sPort struct {
	Name       string `yaml:"name" json:"name"`
	Port       int32  `yaml:"port" json:"port"`
	TargetPort int32  `yaml:"target_port" json:"target_port"`
	Protocol   string `yaml:"protocol" json:"protocol"`
}

// K8sHealthCheck defines health check configuration
type K8sHealthCheck struct {
	LivenessProbe  K8sProbe `yaml:"liveness_probe" json:"liveness_probe"`
	ReadinessProbe K8sProbe `yaml:"readiness_probe" json:"readiness_probe"`
}

// K8sProbe defines a health check probe
type K8sProbe struct {
	Path                string `yaml:"path" json:"path"`
	Port                int32  `yaml:"port" json:"port"`
	InitialDelaySeconds int32  `yaml:"initial_delay_seconds" json:"initial_delay_seconds"`
	PeriodSeconds       int32  `yaml:"period_seconds" json:"period_seconds"`
	TimeoutSeconds      int32  `yaml:"timeout_seconds" json:"timeout_seconds"`
	FailureThreshold    int32  `yaml:"failure_threshold" json:"failure_threshold"`
}

// K8sVolume defines a volume mount
type K8sVolume struct {
	Name      string `yaml:"name" json:"name"`
	MountPath string `yaml:"mount_path" json:"mount_path"`
	Size      string `yaml:"size" json:"size"`
	StorageClass string `yaml:"storage_class" json:"storage_class"`
}

// K8sDeploymentStatus represents deployment status
type K8sDeploymentStatus struct {
	Name               string    `json:"name"`
	Namespace          string    `json:"namespace"`
	Replicas           int32     `json:"replicas"`
	ReadyReplicas      int32     `json:"ready_replicas"`
	AvailableReplicas  int32     `json:"available_replicas"`
	UpdatedReplicas    int32     `json:"updated_replicas"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	Image              string    `json:"image"`
	Conditions         []string  `json:"conditions"`
}

// DefaultK8sConfig returns default Kubernetes configuration
func DefaultK8sConfig() *K8sConfig {
	return &K8sConfig{
		Namespace: "ollamamax",
		Image:     "ollamamax/distributed:latest",
		Replicas:  3,
		Resources: K8sResourceLimits{
			CPURequest:    "500m",
			CPULimit:      "2000m",
			MemoryRequest: "1Gi",
			MemoryLimit:   "4Gi",
		},
		Environment: map[string]string{
			"OLLAMAMAX_MODE": "distributed",
		},
		Labels: map[string]string{
			"app":     "ollamamax",
			"version": "v1.0.0",
		},
		ServiceType: "ClusterIP",
		Ports: []K8sPort{
			{
				Name:       "api",
				Port:       8080,
				TargetPort: 8080,
				Protocol:   "TCP",
			},
			{
				Name:       "web",
				Port:       8081,
				TargetPort: 8081,
				Protocol:   "TCP",
			},
			{
				Name:       "p2p",
				Port:       4001,
				TargetPort: 4001,
				Protocol:   "TCP",
			},
		},
		HealthCheck: K8sHealthCheck{
			LivenessProbe: K8sProbe{
				Path:                "/health",
				Port:                8080,
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
				FailureThreshold:    3,
			},
			ReadinessProbe: K8sProbe{
				Path:                "/health/ready",
				Port:                8080,
				InitialDelaySeconds: 5,
				PeriodSeconds:       5,
				TimeoutSeconds:      3,
				FailureThreshold:    3,
			},
		},
	}
}

// NewKubernetesManager creates a new Kubernetes manager
func NewKubernetesManager(config *K8sConfig, logger *slog.Logger) (*KubernetesManager, error) {
	if config == nil {
		config = DefaultK8sConfig()
	}

	var k8sConfig *rest.Config
	var err error

	if config.Kubeconfig != "" {
		// Load from kubeconfig file
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	} else {
		// Load from in-cluster config
		k8sConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	client, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	return &KubernetesManager{
		client:    client,
		logger:    logger,
		config:    config,
		namespace: namespace,
	}, nil
}

// Deploy deploys to Kubernetes
func (km *KubernetesManager) Deploy(ctx context.Context, name string, overrides *K8sConfig) (*K8sDeploymentStatus, error) {
	config := km.config
	if overrides != nil {
		config = km.mergeK8sConfigs(config, overrides)
	}

	km.logger.Info("Starting Kubernetes deployment", "name", name, "namespace", km.namespace)

	// Create namespace if it doesn't exist
	if err := km.ensureNamespace(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Create deployment
	if err := km.createDeployment(ctx, name, config); err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	// Create service
	if err := km.createService(ctx, name, config); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	// Wait for deployment to be ready
	if err := km.waitForDeployment(ctx, name, 5*time.Minute); err != nil {
		km.logger.Warn("Deployment not ready within timeout", "name", name, "error", err)
	}

	status, err := km.getDeploymentStatus(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment status: %w", err)
	}

	km.logger.Info("Kubernetes deployment completed", "name", name, "replicas", status.ReadyReplicas)
	return status, nil
}

// Scale scales a deployment
func (km *KubernetesManager) Scale(ctx context.Context, name string, replicas int32) error {
	km.logger.Info("Scaling deployment", "name", name, "replicas", replicas)

	deployment, err := km.client.AppsV1().Deployments(km.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = &replicas

	_, err = km.client.AppsV1().Deployments(km.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	// Wait for scaling to complete
	if err := km.waitForDeployment(ctx, name, 3*time.Minute); err != nil {
		km.logger.Warn("Scaling not completed within timeout", "name", name, "error", err)
	}

	km.logger.Info("Deployment scaled successfully", "name", name, "replicas", replicas)
	return nil
}

// Delete removes a deployment
func (km *KubernetesManager) Delete(ctx context.Context, name string) error {
	km.logger.Info("Deleting deployment", "name", name)

	// Delete deployment
	err := km.client.AppsV1().Deployments(km.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		km.logger.Warn("Failed to delete deployment", "name", name, "error", err)
	}

	// Delete service
	err = km.client.CoreV1().Services(km.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		km.logger.Warn("Failed to delete service", "name", name, "error", err)
	}

	km.logger.Info("Deployment deleted", "name", name)
	return nil
}

// GetStatus returns deployment status
func (km *KubernetesManager) GetStatus(ctx context.Context, name string) (*K8sDeploymentStatus, error) {
	return km.getDeploymentStatus(ctx, name)
}

// ListDeployments lists all deployments
func (km *KubernetesManager) ListDeployments(ctx context.Context) ([]*K8sDeploymentStatus, error) {
	deployments, err := km.client.AppsV1().Deployments(km.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=ollamamax",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var statuses []*K8sDeploymentStatus
	for _, deployment := range deployments.Items {
		status := km.convertToStatus(&deployment)
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// Private helper methods

func (km *KubernetesManager) ensureNamespace(ctx context.Context) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: km.namespace,
		},
	}

	_, err := km.client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		// Ignore if namespace already exists
		km.logger.Debug("Namespace creation result", "namespace", km.namespace, "error", err)
	}

	return nil
}

func (km *KubernetesManager) createDeployment(ctx context.Context, name string, config *K8sConfig) error {
	// Convert environment map to env vars
	var envVars []corev1.EnvVar
	for key, value := range config.Environment {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// Convert ports to container ports
	var containerPorts []corev1.ContainerPort
	for _, port := range config.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort,
			Protocol:      corev1.Protocol(port.Protocol),
		})
	}

	// Create deployment spec
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: km.namespace,
			Labels:    config.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &config.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "ollamamax",
					"name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "ollamamax",
						"name": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ollamamax",
							Image: config.Image,
							Ports: containerPorts,
							Env:   envVars,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: config.HealthCheck.LivenessProbe.Path,
										Port: intstr.FromInt32(config.HealthCheck.LivenessProbe.Port),
									},
								},
								InitialDelaySeconds: config.HealthCheck.LivenessProbe.InitialDelaySeconds,
								PeriodSeconds:       config.HealthCheck.LivenessProbe.PeriodSeconds,
								TimeoutSeconds:      config.HealthCheck.LivenessProbe.TimeoutSeconds,
								FailureThreshold:    config.HealthCheck.LivenessProbe.FailureThreshold,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: config.HealthCheck.ReadinessProbe.Path,
										Port: intstr.FromInt32(config.HealthCheck.ReadinessProbe.Port),
									},
								},
								InitialDelaySeconds: config.HealthCheck.ReadinessProbe.InitialDelaySeconds,
								PeriodSeconds:       config.HealthCheck.ReadinessProbe.PeriodSeconds,
								TimeoutSeconds:      config.HealthCheck.ReadinessProbe.TimeoutSeconds,
								FailureThreshold:    config.HealthCheck.ReadinessProbe.FailureThreshold,
							},
						},
					},
				},
			},
		},
	}

	_, err := km.client.AppsV1().Deployments(km.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (km *KubernetesManager) createService(ctx context.Context, name string, config *K8sConfig) error {
	// Convert ports to service ports
	var servicePorts []corev1.ServicePort
	for _, port := range config.Ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.FromInt32(port.TargetPort),
			Protocol:   corev1.Protocol(port.Protocol),
		})
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: km.namespace,
			Labels:    config.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "ollamamax",
				"name": name,
			},
			Ports: servicePorts,
			Type:  corev1.ServiceType(config.ServiceType),
		},
	}

	_, err := km.client.CoreV1().Services(km.namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func (km *KubernetesManager) waitForDeployment(ctx context.Context, name string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := km.getDeploymentStatus(ctx, name)
			if err != nil {
				continue
			}

			if status.ReadyReplicas == status.Replicas && status.Status == "Running" {
				return nil
			}
		}
	}
}

func (km *KubernetesManager) getDeploymentStatus(ctx context.Context, name string) (*K8sDeploymentStatus, error) {
	deployment, err := km.client.AppsV1().Deployments(km.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return km.convertToStatus(deployment), nil
}

func (km *KubernetesManager) convertToStatus(deployment *appsv1.Deployment) *K8sDeploymentStatus {
	status := "Unknown"
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		status = "Running"
	} else if deployment.Status.ReadyReplicas > 0 {
		status = "Partial"
	} else {
		status = "Starting"
	}

	var conditions []string
	for _, cond := range deployment.Status.Conditions {
		if cond.Status == corev1.ConditionTrue {
			conditions = append(conditions, string(cond.Type))
		}
	}

	var image string
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		image = deployment.Spec.Template.Spec.Containers[0].Image
	}

	return &K8sDeploymentStatus{
		Name:               deployment.Name,
		Namespace:          deployment.Namespace,
		Replicas:           *deployment.Spec.Replicas,
		ReadyReplicas:      deployment.Status.ReadyReplicas,
		AvailableReplicas:  deployment.Status.AvailableReplicas,
		UpdatedReplicas:    deployment.Status.UpdatedReplicas,
		Status:             status,
		CreatedAt:          deployment.CreationTimestamp.Time,
		Image:              image,
		Conditions:         conditions,
	}
}

func (km *KubernetesManager) mergeK8sConfigs(base, override *K8sConfig) *K8sConfig {
	merged := *base

	if override.Image != "" {
		merged.Image = override.Image
	}
	if override.Replicas > 0 {
		merged.Replicas = override.Replicas
	}
	if len(override.Environment) > 0 {
		merged.Environment = override.Environment
	}
	if len(override.Labels) > 0 {
		merged.Labels = override.Labels
	}
	if override.ServiceType != "" {
		merged.ServiceType = override.ServiceType
	}

	return &merged
}