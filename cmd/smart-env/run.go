package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/smartenv/smart-env/internal/analyzer"
	"github.com/smartenv/smart-env/internal/checker"
	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/runner"
	"github.com/smartenv/smart-env/internal/ui"
)

var runCmdFlag bool

var runCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "run",
	Short:        "Spin up the environment (services + guidance)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := filepath.Abs(rootPath)
		res, err := analyzer.Scan(root)
		if err != nil {
			return err
		}

		// env issues?
		if len(res.Diff.Missing) > 0 || len(res.Diff.Empty) > 0 {
			fmt.Println(ui.Warn.Render("Env issues detected. Run 'smart-env fix' first."))
			ok, _ := ui.Confirm("Run fix now?", true)
			if ok {
				// delegate by calling fix command
				fixCmd.RunE(cmd, args)
				res, _ = analyzer.Scan(root)
			}
		}

		// deps
		deps := checker.CheckDependencies(res.ProjectTypes)
		for _, d := range deps {
			if d.Required && !d.Installed {
				return fmt.Errorf("missing dependency: %s", d.Name)
			}
		}

		// services
		if res.HasCompose {
			svcs, _ := checker.ParseComposeServices(res.ComposePath)
			svcs = checker.CheckServices(svcs)
			needStart := false
			for _, s := range svcs {
				if !s.Running {
					needStart = true
				}
			}
			if needStart {
				ok, _ := ui.Confirm("Start background services with docker compose?", true)
				if ok {
					if err := runner.StartServices(res.ComposePath); err != nil {
						return err
					}
					ui.Success("services started")
				}
			} else {
				ui.Success("services already running")
			}
		}

		cmdSuggestion := runner.DetectStartCommand(root, res.ProjectTypes)
		fmt.Println()
		ui.PrintTitle("Environment ready")
		if cmdSuggestion != "" {
			fmt.Printf("Start your app with: %s\n", ui.Info.Render(cmdSuggestion))
		}
		if runCmdFlag {
			// attempt run of first token? just print
			logger.Info("auto-run not executed; use detected command")
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&runCmdFlag, "run", false, "attempt to run detected start command")
	rootCmd.AddCommand(runCmd)
}
