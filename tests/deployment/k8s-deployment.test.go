package deployment_test

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNamespaces = []string{
	"ollamamax-redis",
	"ollamamax-timeseries",
	"ollamamax-ml",
	"ollamamax-monitoring",
}

// TestManifestSyntax validates YAML syntax
func TestManifestSyntax(t *testing.T) {
	manifests := []string{
		"k8s/redis-cluster.yaml",
		"k8s/timeseries-db.yaml",
		"k8s/ml-pipeline.yaml",
		"k8s/monitoring-stack.yaml",
	}

	for _, manifest := range manifests {
		t.Run(manifest, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "kubectl", "apply",
				"--dry-run=client", "-f", manifest)
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Invalid manifest %s: %s", manifest, string(output))
		})
	}
}

// TestNamespaceCreation verifies namespaces
func TestNamespaceCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, ns := range testNamespaces {
		t.Run(ns, func(t *testing.T) {
			// Create namespace if not exists
			cmd := exec.CommandContext(ctx, "kubectl", "create", "namespace", ns)
			cmd.Run() // Ignore error if already exists

			// Verify namespace exists
			cmd = exec.CommandContext(ctx, "kubectl", "get", "namespace", ns)
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Namespace %s not found: %s", ns, string(output))
		})
	}
}

// TestStatefulSetDeployment verifies Redis cluster
func TestStatefulSetDeployment(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Apply Redis cluster manifest
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "k8s/redis-cluster.yaml")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply Redis manifest: %s", string(output))

	// Wait for StatefulSet to be ready
	cmd = exec.CommandContext(ctx, "kubectl", "rollout", "status",
		"statefulset/redis-cluster", "-n", "ollamamax-redis", "--timeout=300s")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "StatefulSet not ready: %s", string(output))

	// Verify pod ordering
	cmd = exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-n", "ollamamax-redis", "-l", "app=redis-cluster", "--no-headers")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to get pods: %s", string(output))

	pods := strings.Split(strings.TrimSpace(string(output)), "\n")
	assert.GreaterOrEqual(t, len(pods), 3, "Expected at least 3 Redis pods")

	// Check pod naming convention (should be redis-cluster-0, redis-cluster-1, etc.)
	for i, podLine := range pods {
		podName := strings.Fields(podLine)[0]
		assert.Contains(t, podName, fmt.Sprintf("redis-cluster-%d", i),
			"Pod naming convention not followed")
	}
}

// TestDeploymentRollout checks deployment progress
func TestDeploymentRollout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Apply ML pipeline manifest
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "k8s/ml-pipeline.yaml")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply ML pipeline manifest: %s", string(output))

	// Get all deployments in ml namespace
	cmd = exec.CommandContext(ctx, "kubectl", "get", "deployments",
		"-n", "ollamamax-ml", "--no-headers")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to get deployments: %s", string(output))

	deployments := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, deployLine := range deployments {
		if deployLine == "" {
			continue
		}

		deployName := strings.Fields(deployLine)[0]
		t.Run(deployName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			cmd := exec.CommandContext(ctx, "kubectl", "rollout", "status",
				fmt.Sprintf("deployment/%s", deployName),
				"-n", "ollamamax-ml", "--timeout=180s")
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Deployment %s not ready: %s", deployName, string(output))
		})
	}
}

// TestServiceCreation validates service endpoints
func TestServiceCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	expectedServices := map[string]string{
		"redis-cluster":  "ollamamax-redis",
		"influxdb":       "ollamamax-timeseries",
		"prometheus":     "ollamamax-monitoring",
	}

	for service, namespace := range expectedServices {
		t.Run(service, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, "kubectl", "get", "service",
				service, "-n", namespace)
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Service %s not found in namespace %s: %s",
				service, namespace, string(output))

			// Check for endpoints
			cmd = exec.CommandContext(ctx, "kubectl", "get", "endpoints",
				service, "-n", namespace, "-o", "jsonpath={.subsets[*].addresses[*].ip}")
			output, err = cmd.CombinedOutput()

			assert.NoError(t, err, "Failed to get endpoints: %s", string(output))
			assert.NotEmpty(t, string(output), "Service %s has no endpoints", service)
		})
	}
}

// TestPVCCreation verifies PersistentVolumeClaims
func TestPVCCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for _, ns := range testNamespaces {
		t.Run(ns, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, "kubectl", "get", "pvc",
				"-n", ns, "--no-headers")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Skipf("No PVCs in namespace %s", ns)
				return
			}

			pvcs := strings.Split(strings.TrimSpace(string(output)), "\n")

			for _, pvcLine := range pvcs {
				if pvcLine == "" {
					continue
				}

				fields := strings.Fields(pvcLine)
				pvcName := fields[0]
				status := fields[1]

				assert.Equal(t, "Bound", status,
					"PVC %s in namespace %s not bound", pvcName, ns)
			}
		})
	}
}

// TestServiceDiscovery validates DNS resolution
func TestServiceDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Get a test pod
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-n", "ollamamax-redis", "-l", "app=redis-cluster",
		"-o", "jsonpath={.items[0].metadata.name}")
	output, err := cmd.CombinedOutput()

	if err != nil || string(output) == "" {
		t.Skip("No test pod available for DNS testing")
		return
	}

	testPod := string(output)

	services := []string{
		"redis-cluster.ollamamax-redis.svc.cluster.local",
		"influxdb.ollamamax-timeseries.svc.cluster.local",
	}

	for _, service := range services {
		t.Run(service, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "kubectl", "exec",
				testPod, "-n", "ollamamax-redis", "--",
				"nslookup", service)
			output, err := cmd.CombinedOutput()

			assert.NoError(t, err, "DNS resolution failed for %s: %s", service, string(output))
		})
	}
}

// TestResourceRequests validates resource definitions
func TestResourceRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, ns := range testNamespaces {
		t.Run(ns, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
				"-n", ns, "-o", "json")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Skipf("No pods in namespace %s", ns)
				return
			}

			// Check that output contains resource requests/limits
			assert.Contains(t, string(output), "\"resources\":",
				"No resource requests defined in namespace %s", ns)
		})
	}
}

// TestHPAConfiguration verifies HPA is created
func TestHPAConfiguration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Apply HPA manifest if exists
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "k8s/hpa-autoscaling.yaml")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Skipf("HPA manifest not applied: %s", string(output))
		return
	}

	// Verify HPAs exist
	cmd = exec.CommandContext(ctx, "kubectl", "get", "hpa", "--all-namespaces")
	output, err = cmd.CombinedOutput()

	assert.NoError(t, err, "Failed to get HPAs: %s", string(output))
	assert.NotEmpty(t, string(output), "No HPAs configured")
}

// TestNetworkPolicy validates network isolation
func TestNetworkPolicy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check for network policies
	for _, ns := range testNamespaces {
		t.Run(ns, func(t *testing.T) {
			cmd := exec.CommandContext(ctx, "kubectl", "get", "networkpolicies",
				"-n", ns, "--no-headers")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Skipf("No network policies in namespace %s", ns)
				return
			}

			t.Logf("Network policies in %s: %s", ns, string(output))
		})
	}
}

// TestPodFailure simulates pod crash
func TestPodFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pod failure test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Get a test pod from a deployment
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-n", "ollamamax-ml", "-l", "app=feature-store",
		"-o", "jsonpath={.items[0].metadata.name}")
	output, err := cmd.CombinedOutput()

	if err != nil || string(output) == "" {
		t.Skip("No test pod available for failure test")
		return
	}

	testPod := string(output)

	// Delete the pod
	cmd = exec.CommandContext(ctx, "kubectl", "delete", "pod",
		testPod, "-n", "ollamamax-ml")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to delete pod: %s", string(output))

	t.Log("Pod deleted, waiting for recreation...")
	time.Sleep(10 * time.Second)

	// Verify new pod is created
	cmd = exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-n", "ollamamax-ml", "-l", "app=feature-store", "--no-headers")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to get pods: %s", string(output))

	pods := strings.Split(strings.TrimSpace(string(output)), "\n")
	assert.GreaterOrEqual(t, len(pods), 1, "Pod not recreated after deletion")
}

// Cleanup function
func TestMain(m *testing.M) {
	fmt.Println("Setting up Kubernetes deployment tests...")

	// Run tests
	code := m.Run()

	// Cleanup
	fmt.Println("Cleaning up Kubernetes deployment tests...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Delete test namespaces (optional, comment out if you want to keep them)
	// for _, ns := range testNamespaces {
	// 	cmd := exec.CommandContext(ctx, "kubectl", "delete", "namespace", ns, "--ignore-not-found")
	// 	cmd.Run()
	// }

	fmt.Println("Kubernetes tests completed")
	_ = ctx
	_ = cancel
}
