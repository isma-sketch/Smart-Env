package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/models"
)

var ignoreDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "target": true,
	"dist": true, "build": true, ".idea": true, ".vscode": true,
	"__pycache__": true, "venv": true, ".venv": true, "bin": true,
	"obj": true, ".next": true, ".nuxt": true,
}

func isIgnoredDir(name string) bool { return ignoreDirs[name] }

// Manifest filenames and their project types.
var manifestTypes = map[string]models.ProjectType{
	"package.json":        models.ProjectNode,
	"go.mod":              models.ProjectGo,
	"Cargo.toml":          models.ProjectRust,
	"requirements.txt":    models.ProjectPython,
	"pyproject.toml":      models.ProjectPython,
	"Pipfile":             models.ProjectPython,
	"pom.xml":             models.ProjectJava,
	"build.gradle":        models.ProjectJava,
	"Gemfile":             models.ProjectRuby,
	"docker-compose.yml":  models.ProjectDocker,
	"docker-compose.yaml": models.ProjectDocker,
	"compose.yml":         models.ProjectDocker,
	"compose.yaml":        models.ProjectDocker,
}

// Scan walks the repository root and returns a ScanResult.
func Scan(root string) (models.ScanResult, error) {
	res := models.ScanResult{Root: root, Diff: models.EnvDiff{Mismatch: map[string]models.Mismatch{}}}
	typesSet := map[models.ProjectType]bool{}
	manifests := []string{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if isIgnoredDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		base := filepath.Base(path)
		if pt, ok := manifestTypes[base]; ok {
			typesSet[pt] = true
			manifests = append(manifests, rel)
			logger.Debug("manifest found: %s -> %s", rel, pt)
		}
		// track env files
		if base == ".env" {
			res.EnvFile = path
		}
		if base == ".env.example" || base == ".env.template" || base == ".env.sample" {
			if res.EnvExample == "" {
				res.EnvExample = path
			}
		}
		return nil
	})
	if err != nil {
		return res, err
	}

	for t := range typesSet {
		res.ProjectTypes = append(res.ProjectTypes, t)
	}
	// sort for deterministic output
	// (simple insertion sort)
	for i := 1; i < len(res.ProjectTypes); i++ {
		for j := i; j > 0 && res.ProjectTypes[j] < res.ProjectTypes[j-1]; j-- {
			res.ProjectTypes[j], res.ProjectTypes[j-1] = res.ProjectTypes[j-1], res.ProjectTypes[j]
		}
	}
	res.Manifests = manifests

	// env diff
	if res.EnvExample != "" {
		example, err := ParseEnv(res.EnvExample)
		if err != nil {
			logger.Warn("cannot parse %s: %v", res.EnvExample, err)
		} else {
			active := map[string]string{}
			if res.EnvFile != "" {
				a, err := ParseEnv(res.EnvFile)
				if err == nil {
					active = a
				}
			}
			diff := CompareEnv(example, active)
			res.Diff = diff
		}
	}

	// compose path
	for _, m := range manifests {
		b := filepath.Base(m)
		if b == "docker-compose.yml" || b == "docker-compose.yaml" || b == "compose.yml" || b == "compose.yaml" {
			res.HasCompose = true
			res.ComposePath = filepath.Join(root, m)
		}
	}

	return res, nil
}

// ParseEnv parses an env file into a map of key->value.
func ParseEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip inline comments? keep simple: split on first =
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// remove surrounding quotes
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		out[key] = val
	}
	return out, sc.Err()
}

// ParseEnvOrdered returns keys in file order plus map.
func ParseEnvOrdered(path string) ([]string, map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	order := []string{}
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		quoted := false
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
				quoted = true
			}
		}
		_ = quoted
		if _, ok := out[key]; !ok {
			order = append(order, key)
		}
		out[key] = val
	}
	return order, out, sc.Err()
}

// CompareEnv compares example and active env maps.
func CompareEnv(example, active map[string]string) models.EnvDiff {
	diff := models.EnvDiff{Mismatch: map[string]models.Mismatch{}}
	for k := range example {
		v, ok := active[k]
		if !ok {
			diff.Missing = append(diff.Missing, k)
			continue
		}
		if strings.TrimSpace(v) == "" {
			diff.Empty = append(diff.Empty, k)
		}
		// type/format check: if example value looks like a typed placeholder
		if mm, isMismatch := checkMismatch(k, example[k], v); isMismatch {
			diff.Mismatch[k] = mm
		}
	}
	for k := range active {
		if _, ok := example[k]; !ok {
			diff.Extra = append(diff.Extra, k)
		}
	}
	sort.Strings(diff.Missing)
	sort.Strings(diff.Empty)
	sort.Strings(diff.Extra)
	return diff
}

func checkMismatch(key, exampleVal, actualVal string) (models.Mismatch, bool) {
	// Only flag obvious format mismatches for known patterns.
	el := strings.ToLower(exampleVal)
	// URL expected
	if strings.HasPrefix(el, "http://") || strings.HasPrefix(el, "https://") {
		if !(strings.HasPrefix(actualVal, "http://") || strings.HasPrefix(actualVal, "https://")) {
			if actualVal != "" {
				return models.Mismatch{Key: key, Expected: "URL", Actual: actualVal}, true
			}
		}
	}
	// bool expected
	el2 := strings.ToLower(strings.TrimSpace(exampleVal))
	if (el2 == "true" || el2 == "false") && actualVal != "" {
		al := strings.ToLower(strings.TrimSpace(actualVal))
		if al != "true" && al != "false" {
			return models.Mismatch{Key: key, Expected: "bool", Actual: actualVal}, true
		}
	}
	// port expected (example is numeric)
	if isNumeric(exampleVal) && !isNumeric(actualVal) && actualVal != "" {
		// example numeric but could be port; only flag if key suggests port
		if strings.Contains(strings.ToUpper(key), "PORT") {
			return models.Mismatch{Key: key, Expected: "number", Actual: actualVal}, true
		}
	}
	return models.Mismatch{}, false
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsSecretKey reports whether a key likely holds a secret.
func IsSecretKey(key string) bool {
	u := strings.ToUpper(key)
	for _, kw := range []string{"SECRET", "KEY", "PASSWORD", "PASS", "JWT", "TOKEN", "API_KEY", "PRIVATE"} {
		if strings.Contains(u, kw) {
			return true
		}
	}
	return false
}

// ExampleValue returns example value (may be placeholder).
func ExampleValue(m map[string]string, key string) string { return m[key] }

// FormatMismatch helper for display.
func FormatMismatch(m models.Mismatch) string {
	return fmt.Sprintf("%s (expected %s, got %q)", m.Key, m.Expected, m.Actual)
}
