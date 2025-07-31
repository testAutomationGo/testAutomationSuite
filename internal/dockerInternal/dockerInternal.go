package dockerInternal

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Container struct {
	Name     string
	ID       string
	Image    string
	Status   string
	isActive bool
}

type Client struct {
	containers map[string]*Container
	networks   map[string]string
}

type RunOptions struct {
	Name         string
	Image        string
	Detached     bool
	Capabilities []string
	Volumes      []string
	Ports        []string
	Environment  []string
	Command      []string
	Interactive  bool
	TTY          bool
	Remove       bool
	Network      string
}

func NewClient() *Client {
	return &Client{
		containers: make(map[string]*Container),
		networks:   make(map[string]string),
	}
}

func (c *Client) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func (c *Client) CreateNetwork(networkName string) error {
	if c.NetworkExists(networkName) {
		return fmt.Errorf("network %s already exists in client tracking", networkName)
	}

	existingOutput, _ := c.runCommand("docker", "network", "ls", "--filter", fmt.Sprintf("name=^%s$", networkName), "--format", "{{.Name}}")
	if strings.TrimSpace(existingOutput) == networkName {
		return fmt.Errorf("network %s already exists in Docker", networkName)
	}

	output, err := c.runCommand("docker", "network", "create", networkName)
	if err != nil {
		return fmt.Errorf("failed to create network %s: %w, output: %s", networkName, err, output)
	}

	networkID := strings.TrimSpace(output)
	if len(networkID) > 12 {
		networkID = networkID[:12]
	}

	c.networks[networkName] = networkID
	return nil
}

func (c *Client) CreateOrReuseNetwork(networkName string) error {
	if c.NetworkExists(networkName) {
		return nil
	}

	existingOutput, _ := c.runCommand("docker", "network", "ls", "--filter", fmt.Sprintf("name=^%s$", networkName), "--format", "{{.ID}}")
	if existingID := strings.TrimSpace(existingOutput); existingID != "" {
		if len(existingID) > 12 {
			existingID = existingID[:12]
		}
		c.networks[networkName] = existingID
		return nil
	}

	return c.CreateNetwork(networkName)
}

func (c *Client) ForceCreateNetwork(networkName string) error {
	existingOutput, _ := c.runCommand("docker", "network", "ls", "--filter", fmt.Sprintf("name=^%s$", networkName), "--format", "{{.Name}}")
	if strings.TrimSpace(existingOutput) == networkName {
		_, _ = c.runCommand("docker", "network", "rm", networkName)
	}

	return c.CreateNetwork(networkName)
}

func (c *Client) RemoveNetwork(networkName string) error {
	_, exists := c.networks[networkName]
	if !exists {
		return fmt.Errorf("network %s not found in client tracking", networkName)
	}

	output, err := c.runCommand("docker", "network", "rm", networkName)
	if err != nil {
		return fmt.Errorf("failed to remove network %s: %w, output: %s", networkName, err, output)
	}

	delete(c.networks, networkName)
	return nil
}

func (c *Client) NetworkExists(networkName string) bool {
	_, exists := c.networks[networkName]
	return exists
}

func (c *Client) ListNetworks() []string {
	networks := make([]string, 0, len(c.networks))
	for name := range c.networks {
		networks = append(networks, name)
	}
	return networks
}

func (c *Client) RunContainer(opts RunOptions) (*Container, error) {
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("container-%d", time.Now().Unix())
	}

	args := []string{"run"}

	if opts.Detached {
		args = append(args, "-d")
	}

	if opts.Interactive {
		args = append(args, "-i")
	}

	if opts.TTY {
		args = append(args, "-t")
	}

	if opts.Remove {
		args = append(args, "--rm")
	}

	if opts.Network != "" {
		args = append(args, "--network", opts.Network)
	}

	args = append(args, "--name", opts.Name)

	for _, cap := range opts.Capabilities {
		args = append(args, "--cap-add", cap)
	}

	for _, vol := range opts.Volumes {
		args = append(args, "-v", vol)
	}

	for _, port := range opts.Ports {
		args = append(args, "-p", port)
	}

	for _, env := range opts.Environment {
		args = append(args, "-e", env)
	}

	args = append(args, opts.Image)
	args = append(args, opts.Command...)

	output, err := c.runCommand("docker", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run container: %w, output: %s", err, output)
	}

	containerID := strings.TrimSpace(output)
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	container := &Container{
		Name:     opts.Name,
		ID:       containerID,
		Image:    opts.Image,
		Status:   "running",
		isActive: true,
	}

	c.containers[opts.Name] = container
	return container, nil
}

func (c *Client) ExecCommand(containerName string, command ...string) (string, error) {
	container, exists := c.containers[containerName]
	if !exists {
		return "", fmt.Errorf("container %s not found", containerName)
	}

	if !container.isActive {
		return "", fmt.Errorf("container %s is not active", containerName)
	}

	args := append([]string{"exec", containerName}, command...)
	return c.runCommand("docker", args...)
}

func (c *Client) ExecInteractive(containerName string, command ...string) error {
	container, exists := c.containers[containerName]
	if !exists {
		return fmt.Errorf("container %s not found", containerName)
	}

	if !container.isActive {
		return fmt.Errorf("container %s is not active", containerName)
	}

	args := append([]string{"exec", "-it", containerName}, command...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	return cmd.Run()
}

func (c *Client) StopContainer(containerName string) error {
	container, exists := c.containers[containerName]
	if !exists {
		return fmt.Errorf("container %s not found", containerName)
	}

	output, err := c.runCommand("docker", "stop", containerName)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w, output: %s", err, output)
	}

	container.Status = "stopped"
	container.isActive = false
	return nil
}

func (c *Client) RemoveContainer(containerName string) error {
	container, exists := c.containers[containerName]
	if !exists {
		return fmt.Errorf("container %s not found", containerName)
	}

	if container.isActive {
		if err := c.StopContainer(containerName); err != nil {
			return fmt.Errorf("failed to stop container before removal: %w", err)
		}
	}

	output, err := c.runCommand("docker", "rm", containerName)
	if err != nil {
		return fmt.Errorf("failed to remove container: %w, output: %s", err, output)
	}

	delete(c.containers, containerName)
	return nil
}

func (c *Client) GetContainer(containerName string) (*Container, error) {
	container, exists := c.containers[containerName]
	if !exists {
		return nil, fmt.Errorf("container %s not found", containerName)
	}
	return container, nil
}

func (c *Client) ListContainers() []*Container {
	containers := make([]*Container, 0, len(c.containers))
	for _, container := range c.containers {
		containers = append(containers, container)
	}
	return containers
}

func (c *Client) ContainerExists(containerName string) bool {
	_, exists := c.containers[containerName]
	return exists
}

func (c *Client) IsContainerRunning(containerName string) bool {
	container, exists := c.containers[containerName]
	if !exists {
		return false
	}
	return container.isActive
}

func (c *Client) GetContainerLogs(containerName string, tail int) (string, error) {
	_, exists := c.containers[containerName]
	if !exists {
		return "", fmt.Errorf("container %s not found", containerName)
	}

	args := []string{"logs"}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	args = append(args, containerName)

	return c.runCommand("docker", args...)
}

func (c *Client) WaitForContainer(containerName string, timeout time.Duration) error {
	container, exists := c.containers[containerName]
	if !exists {
		return fmt.Errorf("container %s not found", containerName)
	}

	start := time.Now()
	for time.Since(start) < timeout {
		output, err := c.runCommand("docker", "inspect", "--format", "{{.State.Status}}", containerName)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		if strings.TrimSpace(output) == "running" {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	container.isActive = false
	return fmt.Errorf("container %s did not start within %v", containerName, timeout)
}

func (c *Client) Cleanup() error {
	var errors []string

	for name := range c.containers {
		if err := c.RemoveContainer(name); err != nil {
			errors = append(errors, fmt.Sprintf("failed to cleanup container %s: %v", name, err))
		}
	}

	for name := range c.networks {
		if err := c.RemoveNetwork(name); err != nil {
			errors = append(errors, fmt.Sprintf("failed to cleanup network %s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func GlobalCleanupTestResources(testCasePrefix string) error {
	cmd := exec.Command("docker", "container", "ls", "-a", "--filter", fmt.Sprintf("name=%s", testCasePrefix), "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err == nil {
		containers := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, container := range containers {
			if strings.TrimSpace(container) != "" {
				exec.Command("docker", "rm", "-f", container).Run()
			}
		}
	}

	cmd = exec.Command("docker", "network", "ls", "--filter", fmt.Sprintf("name=%s", testCasePrefix), "--format", "{{.Name}}")
	output, err = cmd.CombinedOutput()
	if err == nil {
		networks := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, network := range networks {
			if strings.TrimSpace(network) != "" && network != "bridge" && network != "host" && network != "none" {
				exec.Command("docker", "network", "rm", network).Run()
			}
		}
	}

	return nil
}

func (c *Client) PullImage(image string) error {
	output, err := c.runCommand("docker", "pull", image)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w, output: %s", image, err, output)
	}
	return nil
}

func (c *Client) CopyToContainer(containerName, srcPath, destPath string) error {
	container, exists := c.containers[containerName]
	if !exists {
		return fmt.Errorf("container %s not found", containerName)
	}

	if !container.isActive {
		return fmt.Errorf("container %s is not active", containerName)
	}

	output, err := c.runCommand("docker", "cp", srcPath, containerName+":"+destPath)
	if err != nil {
		return fmt.Errorf("failed to copy to container: %w, output: %s", err, output)
	}

	return nil
}

func (c *Client) CopyFromContainer(containerName, srcPath, destPath string) error {
	_, exists := c.containers[containerName]
	if !exists {
		return fmt.Errorf("container %s not found", containerName)
	}

	output, err := c.runCommand("docker", "cp", containerName+":"+srcPath, destPath)
	if err != nil {
		return fmt.Errorf("failed to copy from container: %w, output: %s", err, output)
	}

	return nil
}
