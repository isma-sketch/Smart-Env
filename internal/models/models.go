package models

// ProjectType represents a detected project stack.
type ProjectType string

const (
	ProjectNode   ProjectType = "node"
	ProjectGo     ProjectType = "go"
	ProjectRust   ProjectType = "rust"
	ProjectPython ProjectType = "python"
	ProjectJava   ProjectType = "java"
	ProjectRuby   ProjectType = "ruby"
	ProjectDocker ProjectType = "docker"
)

// EnvEntry holds a parsed environment variable.
type EnvEntry struct {
	Key    string
	Value  string
	Quoted bool
}

// EnvDiff captures differences between .env.example and .env.
type EnvDiff struct {
	Missing  []string            // keys in example but not in .env
	Empty    []string            // keys present but empty
	Extra    []string            // keys in .env but not in example
	Mismatch map[string]Mismatch // type/format mismatches
}

// Mismatch describes a value that doesn't match expected format.
type Mismatch struct {
	Key      string
	Expected string
	Actual   string
}

// Dependency is a host CLI tool requirement.
type Dependency struct {
	Name      string
	Command   string
	Args      []string
	Installed bool
	Version   string
	Required  bool
	Project   ProjectType
}

// Service is a background service declared in docker-compose.
type Service struct {
	Name          string
	Image         string
	Ports         []string
	HostPorts     []string
	Running       bool
	ContainerName string
	Known         bool // known DB/cache/etc
}

// ScanResult aggregates all findings.
type ScanResult struct {
	Root         string
	ProjectTypes []ProjectType
	Manifests    []string
	EnvExample   string
	EnvFile      string
	Diff         EnvDiff
	Dependencies []Dependency
	Services     []Service
	HasCompose   bool
	ComposePath  string
}
