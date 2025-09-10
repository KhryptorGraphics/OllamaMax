package autoscaling

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// KubernetesExecutor implements ScalingExecutor for Kubernetes
type KubernetesExecutor struct {
	config *KubernetesConfig

	// Kubernetes clients
	client        kubernetes.Interface
	metricsClient metricsclientset.Interface

	// Current state
	namespace  string
	deployment string
}

// KubernetesConfig holds Kubernetes-specific configuration
type KubernetesConfig struct {
	// Kubernetes connection
	KubeConfig string `yaml:"kube_config"`
	Namespace  string `yaml:"namespace"`
	Deployment string `yaml:"deployment"`

	// Scaling settings
	ScaleTimeout time.Duration `yaml:"scale_timeout"`
	MinReplicas  int32         `yaml:"min_replicas"`
	MaxReplicas  int32         `yaml:"max_replicas"`

	// Resource requests/limits
	CPURequest    string `yaml:"cpu_request"`
	MemoryRequest string `yaml:"memory_request"`
	CPULimit      string `yaml:"cpu_limit"`
	MemoryLimit   string `yaml:"memory_limit"`
}

// DefaultKubernetesConfig returns default Kubernetes configuration
func DefaultKubernetesConfig() *KubernetesConfig {
	return &KubernetesConfig{
		KubeConfig:    "", // Use in-cluster config
		Namespace:     "default",
		Deployment:    "ollama-distributed",
		ScaleTimeout:  5 * time.Minute,
		MinReplicas:   1,
		MaxReplicas:   10,
		CPURequest:    "500m",
		MemoryRequest: "1Gi",
		CPULimit:      "2000m",
		MemoryLimit:   "4Gi",
	}
}

// getKubernetesConfig creates a Kubernetes client configuration
func getKubernetesConfig(kubeConfigPath string) (*rest.Config, error) {
	// Try in-cluster configuration first
	if kubeConfigPath == "" {
		config, err := rest.InClusterConfig()
		if err == nil {
			return config, nil
		}
		fmt.Printf("Failed to use in-cluster config, trying out-of-cluster: %v\n", err)
	}

	// Use out-of-cluster configuration
	if kubeConfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeConfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	return config, nil
}

// NewKubernetesExecutor creates a new Kubernetes scaling executor
func NewKubernetesExecutor(config *KubernetesConfig) (*KubernetesExecutor, error) {
	if config == nil {
		config = DefaultKubernetesConfig()
	}

	// Initialize Kubernetes client configuration
	kubeConfig, err := getKubernetesConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	// Create the Kubernetes client
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create the metrics client
	metricsClient, err := metricsclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	executor := &KubernetesExecutor{
		config:        config,
		client:        client,
		metricsClient: metricsClient,
		namespace:     config.Namespace,
		deployment:    config.Deployment,
	}

	return executor, nil
}

// ScaleUp scales up the deployment to the specified number of replicas
func (ke *KubernetesExecutor) ScaleUp(replicas int) error {
	fmt.Printf("Kubernetes: Scaling up deployment %s/%s to %d replicas\n",
		ke.namespace, ke.deployment, replicas)

	ctx, cancel := context.WithTimeout(context.Background(), ke.config.ScaleTimeout)
	defer cancel()

	// Get the current deployment
	deployment, err := ke.client.AppsV1().Deployments(ke.namespace).Get(ctx, ke.deployment, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s/%s: %w", ke.namespace, ke.deployment, err)
	}

	// Check if scaling is needed
	if deployment.Spec.Replicas != nil && int(*deployment.Spec.Replicas) == replicas {
		fmt.Printf("Deployment %s/%s already has %d replicas\n", ke.namespace, ke.deployment, replicas)
		return nil
	}

	// Update replica count
	replicaCount := int32(replicas)
	deployment.Spec.Replicas = &replicaCount

	_, err = ke.client.AppsV1().Deployments(ke.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment %s/%s: %w", ke.namespace, ke.deployment, err)
	}

	fmt.Printf("Successfully scaled up deployment %s/%s to %d replicas\n", ke.namespace, ke.deployment, replicas)
	return nil
}

// ScaleDown scales down the deployment to the specified number of replicas
func (ke *KubernetesExecutor) ScaleDown(replicas int) error {
	fmt.Printf("Kubernetes: Scaling down deployment %s/%s to %d replicas\n",
		ke.namespace, ke.deployment, replicas)

	ctx, cancel := context.WithTimeout(context.Background(), ke.config.ScaleTimeout)
	defer cancel()

	// Get the current deployment
	deployment, err := ke.client.AppsV1().Deployments(ke.namespace).Get(ctx, ke.deployment, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s/%s: %w", ke.namespace, ke.deployment, err)
	}

	// Check if scaling is needed
	if deployment.Spec.Replicas != nil && int(*deployment.Spec.Replicas) == replicas {
		fmt.Printf("Deployment %s/%s already has %d replicas\n", ke.namespace, ke.deployment, replicas)
		return nil
	}

	// Update replica count with graceful shutdown consideration
	replicaCount := int32(replicas)
	deployment.Spec.Replicas = &replicaCount

	// Ensure graceful termination period is set
	if deployment.Spec.Template.Spec.TerminationGracePeriodSeconds == nil {
		gracePeriod := int64(30) // 30 seconds default
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &gracePeriod
	}

	_, err = ke.client.AppsV1().Deployments(ke.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment %s/%s: %w", ke.namespace, ke.deployment, err)
	}

	fmt.Printf("Successfully scaled down deployment %s/%s to %d replicas\n", ke.namespace, ke.deployment, replicas)
	return nil
}

// GetCurrentReplicas returns the current number of replicas
func (ke *KubernetesExecutor) GetCurrentReplicas() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deployment, err := ke.client.AppsV1().Deployments(ke.namespace).Get(ctx, ke.deployment, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get deployment %s/%s: %w", ke.namespace, ke.deployment, err)
	}

	if deployment.Spec.Replicas == nil {
		return 1, nil
	}

	return int(*deployment.Spec.Replicas), nil
}

// KubernetesMetricsCollector implements MetricsCollector for Kubernetes
type KubernetesMetricsCollector struct {
	config *KubernetesConfig

	// Kubernetes clients
	client        kubernetes.Interface
	metricsClient metricsclientset.Interface
}

// NewKubernetesMetricsCollector creates a new Kubernetes metrics collector
func NewKubernetesMetricsCollector(config *KubernetesConfig) (*KubernetesMetricsCollector, error) {
	if config == nil {
		config = DefaultKubernetesConfig()
	}

	// Initialize Kubernetes client configuration
	kubeConfig, err := getKubernetesConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	// Create the Kubernetes client
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create the metrics client
	metricsClient, err := metricsclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	collector := &KubernetesMetricsCollector{
		config:        config,
		client:        client,
		metricsClient: metricsClient,
	}

	return collector, nil
}

// GetCPUUtilization returns current CPU utilization percentage
func (kmc *KubernetesMetricsCollector) GetCPUUtilization() float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get pod metrics from metrics server
	podMetrics, err := kmc.metricsClient.MetricsV1beta1().PodMetricses(kmc.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kmc.config.Deployment),
	})
	if err != nil {
		fmt.Printf("Failed to get pod metrics: %v\n", err)
		return 0
	}

	if len(podMetrics.Items) == 0 {
		fmt.Printf("No pod metrics found for deployment %s\n", kmc.config.Deployment)
		return 0
	}

	var totalCPUUsage, totalCPURequest int64

	for _, podMetric := range podMetrics.Items {
		for _, container := range podMetric.Containers {
			cpuUsage := container.Usage.Cpu().MilliValue()
			totalCPUUsage += cpuUsage

			// Parse CPU request from config
			cpuRequest, err := resource.ParseQuantity(kmc.config.CPURequest)
			if err != nil {
				fmt.Printf("Failed to parse CPU request: %v\n", err)
				continue
			}
			totalCPURequest += cpuRequest.MilliValue()
		}
	}

	if totalCPURequest == 0 {
		return 0
	}

	utilization := (float64(totalCPUUsage) / float64(totalCPURequest)) * 100
	return utilization
}

// GetMemoryUtilization returns current memory utilization percentage
func (kmc *KubernetesMetricsCollector) GetMemoryUtilization() float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get pod metrics from metrics server
	podMetrics, err := kmc.metricsClient.MetricsV1beta1().PodMetricses(kmc.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kmc.config.Deployment),
	})
	if err != nil {
		fmt.Printf("Failed to get pod metrics: %v\n", err)
		return 0
	}

	if len(podMetrics.Items) == 0 {
		fmt.Printf("No pod metrics found for deployment %s\n", kmc.config.Deployment)
		return 0
	}

	var totalMemoryUsage, totalMemoryRequest int64

	for _, podMetric := range podMetrics.Items {
		for _, container := range podMetric.Containers {
			memoryUsage := container.Usage.Memory().Value()
			totalMemoryUsage += memoryUsage

			// Parse memory request from config
			memoryRequest, err := resource.ParseQuantity(kmc.config.MemoryRequest)
			if err != nil {
				fmt.Printf("Failed to parse memory request: %v\n", err)
				continue
			}
			totalMemoryRequest += memoryRequest.Value()
		}
	}

	if totalMemoryRequest == 0 {
		return 0
	}

	utilization := (float64(totalMemoryUsage) / float64(totalMemoryRequest)) * 100
	return utilization
}

// GetQueueSize returns current queue size from custom metrics
func (kmc *KubernetesMetricsCollector) GetQueueSize() int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to get queue size from pod annotations
	pods, err := kmc.client.CoreV1().Pods(kmc.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kmc.config.Deployment),
	})
	if err != nil {
		fmt.Printf("Failed to get pods: %v\n", err)
		return 0
	}

	totalQueueSize := 0
	for _, pod := range pods.Items {
		if queueSizeStr, ok := pod.Annotations["metrics/queue-size"]; ok {
			if queueSize, err := strconv.Atoi(queueSizeStr); err == nil {
				totalQueueSize += queueSize
			}
		}
	}

	return totalQueueSize
}

// GetResponseTime returns average response time from custom metrics
func (kmc *KubernetesMetricsCollector) GetResponseTime() time.Duration {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to get response time from pod annotations
	pods, err := kmc.client.CoreV1().Pods(kmc.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kmc.config.Deployment),
	})
	if err != nil {
		fmt.Printf("Failed to get pods: %v\n", err)
		return 0
	}

	var totalResponseTime time.Duration
	count := 0

	for _, pod := range pods.Items {
		if responseTimeStr, ok := pod.Annotations["metrics/response-time-ms"]; ok {
			if responseTimeMs, err := strconv.Atoi(responseTimeStr); err == nil {
				totalResponseTime += time.Duration(responseTimeMs) * time.Millisecond
				count++
			}
		}
	}

	if count == 0 {
		return 0
	}

	return totalResponseTime / time.Duration(count)
}

// GetThroughput returns current throughput from custom metrics
func (kmc *KubernetesMetricsCollector) GetThroughput() float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to get throughput from pod annotations
	pods, err := kmc.client.CoreV1().Pods(kmc.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kmc.config.Deployment),
	})
	if err != nil {
		fmt.Printf("Failed to get pods: %v\n", err)
		return 0
	}

	totalThroughput := 0.0
	for _, pod := range pods.Items {
		if throughputStr, ok := pod.Annotations["metrics/throughput-rps"]; ok {
			if throughput, err := strconv.ParseFloat(throughputStr, 64); err == nil {
				totalThroughput += throughput
			}
		}
	}

	return totalThroughput
}

// GetActiveConnections returns number of active connections from custom metrics
func (kmc *KubernetesMetricsCollector) GetActiveConnections() int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to get active connections from pod annotations
	pods, err := kmc.client.CoreV1().Pods(kmc.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kmc.config.Deployment),
	})
	if err != nil {
		fmt.Printf("Failed to get pods: %v\n", err)
		return 0
	}

	totalConnections := 0
	for _, pod := range pods.Items {
		if connectionsStr, ok := pod.Annotations["metrics/active-connections"]; ok {
			if connections, err := strconv.Atoi(connectionsStr); err == nil {
				totalConnections += connections
			}
		}
	}

	return totalConnections
}

// HorizontalPodAutoscaler represents a Kubernetes HPA configuration
type HorizontalPodAutoscaler struct {
	Name      string
	Namespace string
	Target    HPATarget
	Metrics   []HPAMetric
	Behavior  *HPABehavior
}

// HPATarget represents the target resource for scaling
type HPATarget struct {
	APIVersion string
	Kind       string
	Name       string
}

// HPAMetric represents a metric for HPA
type HPAMetric struct {
	Type     string
	Resource *HPAResourceMetric
	Pods     *HPAPodsMetric
	Object   *HPAObjectMetric
}

// HPAResourceMetric represents a resource-based metric
type HPAResourceMetric struct {
	Name   string
	Target HPAMetricTarget
}

// HPAPodsMetric represents a pods-based metric
type HPAPodsMetric struct {
	Metric HPAMetricIdentifier
	Target HPAMetricTarget
}

// HPAObjectMetric represents an object-based metric
type HPAObjectMetric struct {
	DescribedObject HPAObjectReference
	Target          HPAMetricTarget
	Metric          HPAMetricIdentifier
}

// HPAMetricTarget represents a metric target
type HPAMetricTarget struct {
	Type               string
	Value              *int64
	AverageValue       *int64
	AverageUtilization *int32
}

// HPAMetricIdentifier identifies a metric
type HPAMetricIdentifier struct {
	Name     string
	Selector map[string]string
}

// HPAObjectReference references a Kubernetes object
type HPAObjectReference struct {
	APIVersion string
	Kind       string
	Name       string
}

// HPABehavior defines scaling behavior
type HPABehavior struct {
	ScaleUp   *HPAScalingRules
	ScaleDown *HPAScalingRules
}

// HPAScalingRules defines scaling rules
type HPAScalingRules struct {
	StabilizationWindowSeconds *int32
	SelectPolicy               *string
	Policies                   []HPAScalingPolicy
}

// HPAScalingPolicy defines a scaling policy
type HPAScalingPolicy struct {
	Type          string
	Value         int32
	PeriodSeconds int32
}

// convertMetrics converts our HPA metrics to Kubernetes autoscaling metrics
func convertMetrics(metrics []HPAMetric) []autoscalingv2.MetricSpec {
	k8sMetrics := make([]autoscalingv2.MetricSpec, 0, len(metrics))
	
	for _, metric := range metrics {
		k8sMetric := autoscalingv2.MetricSpec{
			Type: autoscalingv2.MetricSourceType(metric.Type),
		}
		
		switch metric.Type {
		case "Resource":
			if metric.Resource != nil {
				k8sMetric.Resource = &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceName(metric.Resource.Name),
					Target: autoscalingv2.MetricTarget{
						Type: autoscalingv2.MetricTargetType(metric.Resource.Target.Type),
					},
				}
				
				if metric.Resource.Target.AverageUtilization != nil {
					k8sMetric.Resource.Target.AverageUtilization = metric.Resource.Target.AverageUtilization
				}
				if metric.Resource.Target.AverageValue != nil {
					avgValue := resource.NewMilliQuantity(*metric.Resource.Target.AverageValue, resource.DecimalSI)
					k8sMetric.Resource.Target.AverageValue = avgValue
				}
			}
			
		case "Pods":
			if metric.Pods != nil {
				k8sMetric.Pods = &autoscalingv2.PodsMetricSource{
					Metric: autoscalingv2.MetricIdentifier{
						Name: metric.Pods.Metric.Name,
					},
					Target: autoscalingv2.MetricTarget{
						Type: autoscalingv2.MetricTargetType(metric.Pods.Target.Type),
					},
				}
				
				if metric.Pods.Target.AverageValue != nil {
					avgValue := resource.NewMilliQuantity(*metric.Pods.Target.AverageValue, resource.DecimalSI)
					k8sMetric.Pods.Target.AverageValue = avgValue
				}
			}
			
		case "Object":
			if metric.Object != nil {
				k8sMetric.Object = &autoscalingv2.ObjectMetricSource{
					DescribedObject: autoscalingv2.CrossVersionObjectReference{
						APIVersion: metric.Object.DescribedObject.APIVersion,
						Kind:       metric.Object.DescribedObject.Kind,
						Name:       metric.Object.DescribedObject.Name,
					},
					Metric: autoscalingv2.MetricIdentifier{
						Name: metric.Object.Metric.Name,
					},
					Target: autoscalingv2.MetricTarget{
						Type: autoscalingv2.MetricTargetType(metric.Object.Target.Type),
					},
				}
				
				if metric.Object.Target.Value != nil {
					value := resource.NewMilliQuantity(*metric.Object.Target.Value, resource.DecimalSI)
					k8sMetric.Object.Target.Value = value
				}
			}
		}
		
		k8sMetrics = append(k8sMetrics, k8sMetric)
	}
	
	return k8sMetrics
}

// convertBehavior converts our HPA behavior to Kubernetes autoscaling behavior
func convertBehavior(behavior *HPABehavior) *autoscalingv2.HorizontalPodAutoscalerBehavior {
	if behavior == nil {
		return nil
	}
	
	k8sBehavior := &autoscalingv2.HorizontalPodAutoscalerBehavior{}
	
	if behavior.ScaleUp != nil {
		k8sBehavior.ScaleUp = &autoscalingv2.HPAScalingRules{
			StabilizationWindowSeconds: behavior.ScaleUp.StabilizationWindowSeconds,
			SelectPolicy:               (*autoscalingv2.ScalingPolicySelect)(behavior.ScaleUp.SelectPolicy),
			Policies:                   make([]autoscalingv2.HPAScalingPolicy, 0, len(behavior.ScaleUp.Policies)),
		}
		
		for _, policy := range behavior.ScaleUp.Policies {
			k8sBehavior.ScaleUp.Policies = append(k8sBehavior.ScaleUp.Policies, autoscalingv2.HPAScalingPolicy{
				Type:          autoscalingv2.HPAScalingPolicyType(policy.Type),
				Value:         policy.Value,
				PeriodSeconds: policy.PeriodSeconds,
			})
		}
	}
	
	if behavior.ScaleDown != nil {
		k8sBehavior.ScaleDown = &autoscalingv2.HPAScalingRules{
			StabilizationWindowSeconds: behavior.ScaleDown.StabilizationWindowSeconds,
			SelectPolicy:               (*autoscalingv2.ScalingPolicySelect)(behavior.ScaleDown.SelectPolicy),
			Policies:                   make([]autoscalingv2.HPAScalingPolicy, 0, len(behavior.ScaleDown.Policies)),
		}
		
		for _, policy := range behavior.ScaleDown.Policies {
			k8sBehavior.ScaleDown.Policies = append(k8sBehavior.ScaleDown.Policies, autoscalingv2.HPAScalingPolicy{
				Type:          autoscalingv2.HPAScalingPolicyType(policy.Type),
				Value:         policy.Value,
				PeriodSeconds: policy.PeriodSeconds,
			})
		}
	}
	
	return k8sBehavior
}

// CreateHPA creates a Kubernetes HPA resource
func (ke *KubernetesExecutor) CreateHPA(hpa *HorizontalPodAutoscaler) error {
	fmt.Printf("Creating HPA: %s/%s\n", hpa.Namespace, hpa.Name)

	ctx, cancel := context.WithTimeout(context.Background(), ke.config.ScaleTimeout)
	defer cancel()

	hpaResource := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hpa.Name,
			Namespace: hpa.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: hpa.Target.APIVersion,
				Kind:       hpa.Target.Kind,
				Name:       hpa.Target.Name,
			},
			MinReplicas: &ke.config.MinReplicas,
			MaxReplicas: ke.config.MaxReplicas,
			Metrics:     convertMetrics(hpa.Metrics),
			Behavior:    convertBehavior(hpa.Behavior),
		},
	}

	_, err := ke.client.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Create(ctx, hpaResource, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create HPA %s/%s: %w", hpa.Namespace, hpa.Name, err)
	}

	fmt.Printf("Successfully created HPA: %s/%s\n", hpa.Namespace, hpa.Name)
	return nil
}

// UpdateHPA updates a Kubernetes HPA resource
func (ke *KubernetesExecutor) UpdateHPA(hpa *HorizontalPodAutoscaler) error {
	fmt.Printf("Updating HPA: %s/%s\n", hpa.Namespace, hpa.Name)

	ctx, cancel := context.WithTimeout(context.Background(), ke.config.ScaleTimeout)
	defer cancel()

	// Get the existing HPA
	existingHPA, err := ke.client.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Get(ctx, hpa.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get existing HPA %s/%s: %w", hpa.Namespace, hpa.Name, err)
	}

	// Update the spec
	existingHPA.Spec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
		APIVersion: hpa.Target.APIVersion,
		Kind:       hpa.Target.Kind,
		Name:       hpa.Target.Name,
	}
	existingHPA.Spec.MinReplicas = &ke.config.MinReplicas
	existingHPA.Spec.MaxReplicas = ke.config.MaxReplicas
	existingHPA.Spec.Metrics = convertMetrics(hpa.Metrics)
	existingHPA.Spec.Behavior = convertBehavior(hpa.Behavior)

	_, err = ke.client.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Update(ctx, existingHPA, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update HPA %s/%s: %w", hpa.Namespace, hpa.Name, err)
	}

	fmt.Printf("Successfully updated HPA: %s/%s\n", hpa.Namespace, hpa.Name)
	return nil
}

// DeleteHPA deletes a Kubernetes HPA resource
func (ke *KubernetesExecutor) DeleteHPA(namespace, name string) error {
	fmt.Printf("Deleting HPA: %s/%s\n", namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), ke.config.ScaleTimeout)
	defer cancel()

	err := ke.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete HPA %s/%s: %w", namespace, name, err)
	}

	fmt.Printf("Successfully deleted HPA: %s/%s\n", namespace, name)
	return nil
}

// GetHPAStatus returns the status of an HPA
func (ke *KubernetesExecutor) GetHPAStatus(namespace, name string) (*HPAStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hpa, err := ke.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get HPA %s/%s: %w", namespace, name, err)
	}

	conditions := make([]HPACondition, 0, len(hpa.Status.Conditions))
	for _, condition := range hpa.Status.Conditions {
		conditions = append(conditions, HPACondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	var lastScaleTime time.Time
	if hpa.Status.LastScaleTime != nil {
		lastScaleTime = hpa.Status.LastScaleTime.Time
	}

	return &HPAStatus{
		CurrentReplicas: hpa.Status.CurrentReplicas,
		DesiredReplicas: hpa.Status.DesiredReplicas,
		LastScaleTime:   lastScaleTime,
		Conditions:      conditions,
	}, nil
}

// HPAStatus represents HPA status
type HPAStatus struct {
	CurrentReplicas int32
	DesiredReplicas int32
	LastScaleTime   time.Time
	Conditions      []HPACondition
}

// HPACondition represents an HPA condition
type HPACondition struct {
	Type               string
	Status             string
	LastTransitionTime time.Time
	Reason             string
	Message            string
}