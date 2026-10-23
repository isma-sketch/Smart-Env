package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/smartenv/smart-env/internal/checker"
	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/models"
	"github.com/smartenv/smart-env/internal/ui"
)

// StartServices runs docker compose up -d if compose present.
func StartServices(composePath string) error {
	if composePath == "" {
		return nil
	}
	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger.Debug("running: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		// try docker-compose
		cmd2 := exec.Command("docker-compose", "-f", composePath, "up", "-d")
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err2 := cmd2.Run(); err2 != nil {
			return fmt.Errorf("failed to start services: %v / %v", err, err2)
		}
	}
	return nil
}

// DetectStartCommand suggests a dev/start command.
func DetectStartCommand(root string, types []models.ProjectType) string {
	// Node
	if containsType(types, models.ProjectNode) {
		// check package.json scripts
		return "npm run dev (or npm start)"
	}
	if containsType(types, models.ProjectGo) {
		return "go run ."
	}
	if containsType(types, models.ProjectRust) {
		return "cargo run"
	}
	if containsType(types, models.ProjectPython) {
		return "python -m <module> / flask run"
	}
	return ""
}

func containsType(types []models.ProjectType, t models.ProjectType) bool {
	for _, x := range types {
		if x == t {
			return true
		}
	}
	return false
}

// RunCommand runs a shell command in root.
func RunCommand(root, cmdStr string) error {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}
	c := exec.Command(parts[0], parts[1:]...)
	c.Dir = root
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

// EnsureDepsOk checks all deps installed.
func EnsureDepsOk(deps []models.Dependency) bool {
	ok := true
	for _, d := range deps {
		if d.Required && !d.Installed {
			ok = false
		}
	}
	return ok
}

// PromptStartServices asks user.
func PromptStartServices() (bool, error) {
	return ui.Confirm("Start detected background services with docker compose?", true)
}

var _ = checker.CheckToolInstalled
