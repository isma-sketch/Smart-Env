package checker

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/models"
	"gopkg.in/yaml.v3"
)

// RequiredToolsByType returns tool names needed for a project type.
func RequiredToolsByType(pt models.ProjectType) []string {
	switch pt {
	case models.ProjectNode:
		return []string{"node", "npm"}
	case models.ProjectGo:
		return []string{"go"}
	case models.ProjectRust:
		return []string{"cargo", "rustc"}
	case models.ProjectPython:
		return []string{"python3", "pip3"}
	case models.ProjectJava:
		return []string{"java", "mvn"}
	case models.ProjectRuby:
		return []string{"ruby", "bundle"}
	case models.ProjectDocker:
		return []string{"docker"}
	}
	return nil
}

// CheckDependencies checks host tools for detected project types.
func CheckDependencies(types []models.ProjectType) []models.Dependency {
	seen := map[string]bool{}
	var deps []models.Dependency
	for _, t := range types {
		for _, tool := range RequiredToolsByType(t) {
			if seen[tool] {
				continue
			}
			seen[tool] = true
			d := checkTool(tool, t)
			deps = append(deps, d)
		}
	}
	// docker compose availability if docker present
	if contains(types, models.ProjectDocker) {
		// check docker compose
		d := checkTool("docker compose", models.ProjectDocker)
		d.Name = "docker compose"
		d.Command = "docker"
		d.Args = []string{"compose", "version"}
		// recompute
		d = checkCompose()
		deps = append(deps, d)
	}
	// lockfile-based extras
	return deps
}

func contains(types []models.ProjectType, t models.ProjectType) bool {
	for _, x := range types {
		if x == t {
			return true
		}
	}
	return false
}

func checkTool(name string, pt models.ProjectType) models.Dependency {
	d := models.Dependency{Name: name, Command: name, Required: true, Project: pt}
	// map python3 fallback
	cmd := name
	args := []string{"--version"}
	if name == "node" {
		args = []string{"-v"}
	}
	if name == "go" {
		args = []string{"version"}
	}
	if name == "ruby" {
		args = []string{"--version"}
	}
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		// try fallback for python
		if name == "python3" {
			out, err = exec.Command("python", "--version").CombinedOutput()
			if err == nil {
				d.Command = "python"
				d.Installed = true
				d.Version = strings.TrimSpace(string(out))
				return d
			}
		}
		if name == "pip3" {
			out, err = exec.Command("pip", "--version").CombinedOutput()
			if err == nil {
				d.Command = "pip"
				d.Installed = true
				d.Version = strings.TrimSpace(string(out))
				return d
			}
		}
		d.Installed = false
		logger.Debug("tool %s not found: %v", name, err)
		return d
	}
	d.Installed = true
	d.Version = strings.TrimSpace(string(out))
	// take first line
	if i := strings.Index(d.Version, "\n"); i >= 0 {
		d.Version = d.Version[:i]
	}
	return d
}

func checkCompose() models.Dependency {
	d := models.Dependency{Name: "docker compose", Command: "docker", Args: []string{"compose", "version"}, Required: true, Project: models.ProjectDocker}
	out, err := exec.Command("docker", "compose", "version").CombinedOutput()
	if err != nil {
		// try docker-compose
		out2, err2 := exec.Command("docker-compose", "version").CombinedOutput()
		if err2 != nil {
			d.Installed = false
			return d
		}
		d.Command = "docker-compose"
		d.Args = []string{"version"}
		d.Installed = true
		d.Version = strings.TrimSpace(string(out2))
		if i := strings.Index(d.Version, "\n"); i >= 0 {
			d.Version = d.Version[:i]
		}
		return d
	}
	d.Installed = true
	d.Version = strings.TrimSpace(string(out))
	if i := strings.Index(d.Version, "\n"); i >= 0 {
		d.Version = d.Version[:i]
	}
	return d
}

// ParseComposeServices parses docker-compose file for services.
func ParseComposeServices(path string) ([]models.Service, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	servicesRaw, ok := doc["services"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no services key")
	}
	var services []models.Service
	for name, v := range servicesRaw {
		m, _ := v.(map[string]interface{})
		svc := models.Service{Name: name}
		if img, ok := m["image"].(string); ok {
			svc.Image = img
		}
		if ports, ok := m["ports"].([]interface{}); ok {
			for _, p := range ports {
				ps := fmt.Sprintf("%v", p)
				svc.Ports = append(svc.Ports, ps)
				// host port
				host := ps
				if idx := strings.Index(ps, ":"); idx >= 0 {
					host = ps[:idx]
				}
				// strip protocol
				host = strings.Split(host, "/")[0]
				svc.HostPorts = append(svc.HostPorts, host)
			}
		}
		svc.Known = isKnownService(svc.Image, name)
		services = append(services, svc)
	}
	return services, nil
}

func isKnownService(image, name string) bool {
	l := strings.ToLower(image + " " + name)
	for _, k := range []string{"postgres", "redis", "rabbitmq", "mysql", "mongo", "mariadb", "elasticsearch", "kafka"} {
		if strings.Contains(l, k) {
			return true
		}
	}
	return false
}

// CheckServices updates services with running status.
func CheckServices(services []models.Service) []models.Service {
	for i := range services {
		// container running?
		running := false
		out, err := exec.Command("docker", "ps", "--format", "{{.Names}}").CombinedOutput()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.Contains(strings.ToLower(line), strings.ToLower(services[i].Name)) {
					running = true
					services[i].ContainerName = line
				}
			}
		}
		// port check
		portOpen := false
		for _, hp := range services[i].HostPorts {
			if hp == "" {
				continue
			}
			conn, err := net.DialTimeout("tcp", "127.0.0.1:"+hp, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				portOpen = true
			}
		}
		if len(services[i].HostPorts) == 0 {
			// no host ports; rely on container
			services[i].Running = running
		} else {
			services[i].Running = running || portOpen
		}
	}
	return services
}

// CheckToolInstalled is exported helper.
func CheckToolInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ReadComposeFile helper.
func ReadComposeFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// scan lines helper unused.
func scanLines(s string) []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	return out
}
