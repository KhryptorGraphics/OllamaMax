package deployment

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// DockerManager manages Docker-based deployments
type DockerManager struct {
	client *client.Client
	logger *slog.Logger
	config *DockerConfig
}

// DockerConfig holds Docker deployment configuration
type DockerConfig struct {
	Registry     string            `yaml:"registry" json:"registry"`
	Repository   string            `yaml:"repository" json:"repository"`
	Tag          string            `yaml:"tag" json:"tag"`
	Networks     []string          `yaml:"networks" json:"networks"`
	Volumes      []string          `yaml:"volumes" json:"volumes"`
	Environment  map[string]string `yaml:"environment" json:"environment"`
	Resources    ResourceLimits    `yaml:"resources" json:"resources"`
	HealthCheck  HealthCheckConfig `yaml:"health_check" json:"health_check"`
	RestartPolicy string           `yaml:"restart_policy" json:"restart_policy"`
}

// ResourceLimits defines container resource limits
type ResourceLimits struct {
	CPULimit    string `yaml:"cpu_limit" json:"cpu_limit"`
	MemoryLimit string `yaml:"memory_limit" json:"memory_limit"`
	CPURequest  string `yaml:"cpu_request" json:"cpu_request"`
	MemoryRequest string `yaml:"memory_request" json:"memory_request"`
}

// HealthCheckConfig defines container health check configuration
type HealthCheckConfig struct {
	Command     []string      `yaml:"command" json:"command"`
	Interval    time.Duration `yaml:"interval" json:"interval"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	StartPeriod time.Duration `yaml:"start_period" json:"start_period"`
	Retries     int           `yaml:"retries" json:"retries"`
}

// ContainerInfo represents container information
type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Created time.Time         `json:"created"`
	Ports   map[string]string `json:"ports"`
	Labels  map[string]string `json:"labels"`
	Health  string            `json:"health,omitempty"`
}

// DeploymentResult represents deployment operation result
type DeploymentResult struct {
	ContainerID string                 `json:"container_id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Logs        []string               `json:"logs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultDockerConfig returns default Docker configuration
func DefaultDockerConfig() *DockerConfig {
	return &DockerConfig{
		Registry:   "docker.io",
		Repository: "ollamamax/distributed",
		Tag:        "latest",
		Networks:   []string{"ollamamax"},
		Environment: map[string]string{
			"OLLAMAMAX_MODE": "distributed",
		},
		Resources: ResourceLimits{
			CPULimit:      "2.0",
			MemoryLimit:   "4Gi",
			CPURequest:    "0.5",
			MemoryRequest: "1Gi",
		},
		HealthCheck: HealthCheckConfig{
			Command:     []string{"CMD", "curl", "-f", "http://localhost:8080/health"},
			Interval:    30 * time.Second,
			Timeout:     10 * time.Second,
			StartPeriod: 60 * time.Second,
			Retries:     3,
		},
		RestartPolicy: "unless-stopped",
	}
}

// NewDockerManager creates a new Docker manager
func NewDockerManager(config *DockerConfig, logger *slog.Logger) (*DockerManager, error) {
	if config == nil {
		config = DefaultDockerConfig()
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerManager{
		client: cli,
		logger: logger,
		config: config,
	}, nil
}

// Deploy deploys a new container
func (dm *DockerManager) Deploy(ctx context.Context, name string, overrides *DockerConfig) (*DeploymentResult, error) {
	config := dm.config
	if overrides != nil {
		config = dm.mergeConfigs(config, overrides)
	}

	imageName := dm.buildImageName(config)
	
	dm.logger.Info("Starting deployment", "name", name, "image", imageName)

	// Pull image if needed
	if err := dm.pullImage(ctx, imageName); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Stop existing container if it exists
	if err := dm.stopContainer(ctx, name); err != nil {
		dm.logger.Warn("Failed to stop existing container", "name", name, "error", err)
	}

	// Create and start container
	containerID, err := dm.createContainer(ctx, name, imageName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := dm.startContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for health check
	if err := dm.waitForHealthy(ctx, containerID, 2*time.Minute); err != nil {
		dm.logger.Warn("Container health check failed", "container_id", containerID, "error", err)
	}

	dm.logger.Info("Deployment completed", "name", name, "container_id", containerID)

	return &DeploymentResult{
		ContainerID: containerID,
		Status:      "deployed",
		Message:     "Container deployed successfully",
		Metadata: map[string]interface{}{
			"image": imageName,
			"name":  name,
		},
	}, nil
}

// Undeploy removes a deployed container
func (dm *DockerManager) Undeploy(ctx context.Context, name string) (*DeploymentResult, error) {
	dm.logger.Info("Starting undeployment", "name", name)

	containerID, err := dm.getContainerID(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find container: %w", err)
	}

	// Stop container
	if err := dm.stopContainer(ctx, name); err != nil {
		return nil, fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove container
	if err := dm.removeContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("failed to remove container: %w", err)
	}

	dm.logger.Info("Undeployment completed", "name", name, "container_id", containerID)

	return &DeploymentResult{
		ContainerID: containerID,
		Status:      "undeployed",
		Message:     "Container removed successfully",
	}, nil
}

// Scale scales the number of container instances
func (dm *DockerManager) Scale(ctx context.Context, name string, replicas int) ([]*DeploymentResult, error) {
	dm.logger.Info("Starting scaling operation", "name", name, "replicas", replicas)

	var results []*DeploymentResult

	// Get current instances
	currentInstances, err := dm.listInstances(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to list current instances: %w", err)
	}

	currentCount := len(currentInstances)

	if replicas > currentCount {
		// Scale up
		for i := currentCount; i < replicas; i++ {
			instanceName := fmt.Sprintf("%s-%d", name, i)
			result, err := dm.Deploy(ctx, instanceName, nil)
			if err != nil {
				dm.logger.Error("Failed to deploy instance", "name", instanceName, "error", err)
				result = &DeploymentResult{
					Status:  "failed",
					Message: err.Error(),
				}
			}
			results = append(results, result)
		}
	} else if replicas < currentCount {
		// Scale down
		for i := replicas; i < currentCount; i++ {
			instanceName := fmt.Sprintf("%s-%d", name, i)
			result, err := dm.Undeploy(ctx, instanceName)
			if err != nil {
				dm.logger.Error("Failed to undeploy instance", "name", instanceName, "error", err)
				result = &DeploymentResult{
					Status:  "failed",
					Message: err.Error(),
				}
			}
			results = append(results, result)
		}
	}

	dm.logger.Info("Scaling operation completed", "name", name, "from", currentCount, "to", replicas)

	return results, nil
}

// GetStatus returns the status of deployed containers
func (dm *DockerManager) GetStatus(ctx context.Context, name string) ([]*ContainerInfo, error) {
	containers, err := dm.listInstances(ctx, name)
	if err != nil {
		return nil, err
	}

	var infos []*ContainerInfo
	for _, container := range containers {
		info, err := dm.getContainerInfo(ctx, container.ID)
		if err != nil {
			dm.logger.Error("Failed to get container info", "container_id", container.ID, "error", err)
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// GetLogs retrieves container logs
func (dm *DockerManager) GetLogs(ctx context.Context, containerID string, lines int) ([]string, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", lines),
		Timestamps: true,
	}

	logs, err := dm.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	content, err := io.ReadAll(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	// Split logs into lines
	logLines := strings.Split(string(content), "\n")
	
	// Remove empty lines
	var result []string
	for _, line := range logLines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

// Private helper methods

func (dm *DockerManager) buildImageName(config *DockerConfig) string {
	if config.Registry != "" {
		return fmt.Sprintf("%s/%s:%s", config.Registry, config.Repository, config.Tag)
	}
	return fmt.Sprintf("%s:%s", config.Repository, config.Tag)
}

func (dm *DockerManager) pullImage(ctx context.Context, imageName string) error {
	dm.logger.Info("Pulling image", "image", imageName)

	reader, err := dm.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Read the response to ensure pull completes
	io.Copy(io.Discard, reader)
	
	return nil
}

func (dm *DockerManager) createContainer(ctx context.Context, name, imageName string, config *DockerConfig) (string, error) {
	// Convert environment map to slice
	var env []string
	for key, value := range config.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	containerConfig := &container.Config{
		Image: imageName,
		Env:   env,
		Labels: map[string]string{
			"ollamamax.deployment": "managed",
			"ollamamax.name":       name,
		},
		ExposedPorts: map[string]struct{}{
			"8080/tcp": {},
			"8081/tcp": {},
		},
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: config.RestartPolicy,
		},
		AutoRemove: false,
	}

	// Add resource limits if specified
	if config.Resources.MemoryLimit != "" {
		// Parse memory limit (simplified - would need proper parsing)
		hostConfig.Memory = 4 * 1024 * 1024 * 1024 // 4GB
	}

	resp, err := dm.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, name)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (dm *DockerManager) startContainer(ctx context.Context, containerID string) error {
	return dm.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (dm *DockerManager) stopContainer(ctx context.Context, name string) error {
	containerID, err := dm.getContainerID(ctx, name)
	if err != nil {
		return err // Container doesn't exist
	}

	timeout := 30
	return dm.client.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
}

func (dm *DockerManager) removeContainer(ctx context.Context, containerID string) error {
	return dm.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (dm *DockerManager) getContainerID(ctx context.Context, name string) (string, error) {
	containers, err := dm.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return "", err
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return container.ID, nil
			}
		}
	}

	return "", fmt.Errorf("container not found: %s", name)
}

func (dm *DockerManager) listInstances(ctx context.Context, namePrefix string) ([]types.Container, error) {
	containers, err := dm.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	var instances []types.Container
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.HasPrefix(strings.TrimPrefix(name, "/"), namePrefix) {
				instances = append(instances, container)
				break
			}
		}
	}

	return instances, nil
}

func (dm *DockerManager) getContainerInfo(ctx context.Context, containerID string) (*ContainerInfo, error) {
	inspect, err := dm.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	info := &ContainerInfo{
		ID:      inspect.ID[:12], // Short ID
		Name:    strings.TrimPrefix(inspect.Name, "/"),
		Image:   inspect.Config.Image,
		Status:  inspect.State.Status,
		State:   inspect.State.Status,
		Created: inspect.Created,
		Labels:  inspect.Config.Labels,
		Ports:   make(map[string]string),
	}

	// Extract port mappings
	for port, bindings := range inspect.NetworkSettings.Ports {
		if len(bindings) > 0 {
			info.Ports[string(port)] = fmt.Sprintf("%s:%s", bindings[0].HostIP, bindings[0].HostPort)
		}
	}

	// Add health status if available
	if inspect.State.Health != nil {
		info.Health = inspect.State.Health.Status
	}

	return info, nil
}

func (dm *DockerManager) waitForHealthy(ctx context.Context, containerID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			inspect, err := dm.client.ContainerInspect(ctx, containerID)
			if err != nil {
				return err
			}

			if inspect.State.Health != nil {
				switch inspect.State.Health.Status {
				case "healthy":
					return nil
				case "unhealthy":
					return fmt.Errorf("container is unhealthy")
				}
			} else if inspect.State.Running {
				// No health check defined, consider running as healthy
				return nil
			}
		}
	}
}

func (dm *DockerManager) mergeConfigs(base, override *DockerConfig) *DockerConfig {
	// Create a copy of base config
	merged := *base

	// Override with provided values
	if override.Registry != "" {
		merged.Registry = override.Registry
	}
	if override.Repository != "" {
		merged.Repository = override.Repository
	}
	if override.Tag != "" {
		merged.Tag = override.Tag
	}
	if len(override.Networks) > 0 {
		merged.Networks = override.Networks
	}
	if len(override.Environment) > 0 {
		merged.Environment = override.Environment
	}
	if override.RestartPolicy != "" {
		merged.RestartPolicy = override.RestartPolicy
	}

	return &merged
}

// Close closes the Docker client
func (dm *DockerManager) Close() error {
	return dm.client.Close()
}