package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/smartenv/smart-env/internal/analyzer"
	"github.com/smartenv/smart-env/internal/checker"
	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/models"
	"github.com/smartenv/smart-env/internal/ui"
)

var checkCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "check",
	Short:        "Scan repository and check env/dependencies/services",
	Aliases:      []string{"scan"},
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := filepath.Abs(rootPath)
		res, err := analyzer.Scan(root)
		if err != nil {
			return err
		}
		printScan(res)
		// deps
		deps := checker.CheckDependencies(res.ProjectTypes)
		printDeps(deps)
		// services
		if res.HasCompose {
			svcs, err := checker.ParseComposeServices(res.ComposePath)
			if err != nil {
				logger.Warn("cannot parse compose: %v", err)
			} else {
				svcs = checker.CheckServices(svcs)
				res.Services = svcs
				printServices(svcs)
			}
		}
		// exit code if issues
		issues := countIssues(res, deps)
		if issues > 0 {
			return fmt.Errorf("%d issue(s) found", issues)
		}
		ui.PrintTitle("All checks passed")
		return nil
	},
}

func printScan(res models.ScanResult) {
	ui.PrintTitle("Repository Scan")
	fmt.Printf("Root: %s\n", res.Root)
	fmt.Printf("Project types: %v\n", res.ProjectTypes)
	fmt.Printf("Manifests: %v\n", res.Manifests)
	if res.EnvExample != "" {
		fmt.Printf("Env example: %s\n", res.EnvExample)
	}
	if res.EnvFile != "" {
		fmt.Printf("Env file: %s\n", res.EnvFile)
	}
	fmt.Println()
	fmt.Println(ui.Bold.Render("Env diff:"))
	if len(res.Diff.Missing) == 0 && len(res.Diff.Empty) == 0 && len(res.Diff.Mismatch) == 0 {
		fmt.Println(ui.OK.Render("  no missing/empty/mismatched keys"))
	} else {
		for _, k := range res.Diff.Missing {
			fmt.Printf("  %s %s\n", ui.Fail.Render("MISSING"), k)
		}
		for _, k := range res.Diff.Empty {
			fmt.Printf("  %s %s\n", ui.Warn.Render("EMPTY  "), k)
		}
		for k, m := range res.Diff.Mismatch {
			fmt.Printf("  %s %s (expected %s)\n", ui.Warn.Render("MISMATCH"), k, m.Expected)
		}
	}
	if len(res.Diff.Extra) > 0 {
		fmt.Printf("  %s %v\n", ui.Dim.Render("extra:"), res.Diff.Extra)
	}
	fmt.Println()
}

func printDeps(deps []models.Dependency) {
	fmt.Println(ui.Bold.Render("Dependencies:"))
	for _, d := range deps {
		status := ui.Fail.Render("MISSING")
		if d.Installed {
			status = ui.OK.Render("OK     ")
		}
		fmt.Printf("  %s %-16s %s\n", status, d.Name, d.Version)
	}
	fmt.Println()
}

func printServices(svcs []models.Service) {
	fmt.Println(ui.Bold.Render("Services (docker-compose):"))
	for _, s := range svcs {
		status := ui.Fail.Render("STOPPED")
		if s.Running {
			status = ui.OK.Render("RUNNING")
		}
		ports := ""
		if len(s.HostPorts) > 0 {
			ports = " ports=" + join(s.HostPorts, ",")
		}
		fmt.Printf("  %s %-16s %s%s\n", status, s.Name, s.Image, ports)
	}
	fmt.Println()
}

func join(ss []string, sep string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

func countIssues(res models.ScanResult, deps []models.Dependency) int {
	n := len(res.Diff.Missing) + len(res.Diff.Empty) + len(res.Diff.Mismatch)
	for _, d := range deps {
		if d.Required && !d.Installed {
			n++
		}
	}
	for _, s := range res.Services {
		if !s.Running {
			n++
		}
	}
	_ = sort.Strings
	return n
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
