package autoscaling

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/autoscaling"
)

func TestKubernetesConfig(t *testing.T) {
	config := autoscaling.DefaultKubernetesConfig()
	
	if config == nil {
		t.Fatal("DefaultKubernetesConfig returned nil")
	}
	
	if config.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", config.Namespace)
	}
	
	if config.Deployment != "ollama-distributed" {
		t.Errorf("Expected deployment 'ollama-distributed', got '%s'", config.Deployment)
	}
	
	if config.ScaleTimeout != 5*time.Minute {
		t.Errorf("Expected scale timeout '5m', got '%v'", config.ScaleTimeout)
	}
	
	if config.MinReplicas != 1 {
		t.Errorf("Expected min replicas '1', got '%d'", config.MinReplicas)
	}
	
	if config.MaxReplicas != 10 {
		t.Errorf("Expected max replicas '10', got '%d'", config.MaxReplicas)
	}
}

func TestHPAStructures(t *testing.T) {
	// Test that HPA structures are properly defined
	hpa := &autoscaling.HorizontalPodAutoscaler{
		Name:      "test-hpa",
		Namespace: "default",
		Target: autoscaling.HPATarget{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "test-deployment",
		},
		Metrics: []autoscaling.HPAMetric{
			{
				Type: "Resource",
				Resource: &autoscaling.HPAResourceMetric{
					Name: "cpu",
					Target: autoscaling.HPAMetricTarget{
						Type:               "Utilization",
						AverageUtilization: func() *int32 { v := int32(70); return &v }(),
					},
				},
			},
		},
	}
	
	if hpa.Name != "test-hpa" {
		t.Errorf("Expected HPA name 'test-hpa', got '%s'", hpa.Name)
	}
	
	if len(hpa.Metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(hpa.Metrics))
	}
	
	if hpa.Metrics[0].Type != "Resource" {
		t.Errorf("Expected metric type 'Resource', got '%s'", hpa.Metrics[0].Type)
	}
}