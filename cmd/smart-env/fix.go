package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/smartenv/smart-env/internal/analyzer"
	"github.com/smartenv/smart-env/internal/config"
	"github.com/smartenv/smart-env/internal/generator"
	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/runner"
	"github.com/smartenv/smart-env/internal/ui"
)

var fixCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "fix",
	Short:        "Interactively resolve missing/empty env vars",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := filepath.Abs(rootPath)
		res, err := analyzer.Scan(root)
		if err != nil {
			return err
		}
		if res.EnvExample == "" {
			return fmt.Errorf("no .env.example found")
		}
		exampleOrder, exampleVals, err := config.ReadEnvOrdered(res.EnvExample)
		if err != nil {
			return err
		}
		activeOrder, activeVals, _ := []string{}, map[string]string{}, error(nil)
		if res.EnvFile != "" {
			activeOrder, activeVals, _ = config.ReadEnvOrdered(res.EnvFile)
		}
		_ = activeOrder

		// build final values map starting from active
		values := map[string]string{}
		for k, v := range activeVals {
			values[k] = v
		}
		// extras = keys in active not in example
		extraSet := map[string]bool{}
		for k := range activeVals {
			if _, ok := exampleVals[k]; !ok {
				extraSet[k] = true
			}
		}
		var extras []string
		for k := range extraSet {
			extras = append(extras, k)
		}
		sort.Strings(extras)

		// determine keys needing input: missing or empty
		needsInput := []string{}
		for _, k := range res.Diff.Missing {
			needsInput = append(needsInput, k)
		}
		for _, k := range res.Diff.Empty {
			needsInput = append(needsInput, k)
		}
		// dedup
		seen := map[string]bool{}
		var unique []string
		for _, k := range needsInput {
			if !seen[k] {
				seen[k] = true
				unique = append(unique, k)
			}
		}
		needsInput = unique

		if len(needsInput) == 0 {
			fmt.Println(ui.OK.Render("No missing or empty env vars."))
		} else {
			ui.PrintTitle("Resolving env vars")
			for _, k := range needsInput {
				exVal := exampleVals[k]
				secret := analyzer.IsSecretKey(k)
				gen := generator.IsGeneratable(k)
				if gen {
					action, err := ui.PromptGenerate(k)
					if err != nil {
						if err.Error() == "^C" {
							return err
						}
						logger.Warn("skipping %s: %v", k, err)
						continue
					}
					if action == "generate" {
						values[k] = generator.GenerateFor(k)
						fmt.Printf("  %s %s = %s\n", ui.OK.Render("generated"), k, mask(values[k]))
						continue
					}
					if action == "skip" {
						values[k] = values[k]
						continue
					}
				}
				def := values[k]
				if def == "" {
					def = ""
				}
				v, err := ui.PromptValue(k, exVal, def, secret)
				if err != nil {
					if err.Error() == "^C" {
						return err
					}
					logger.Warn("skipping %s: %v", k, err)
					continue
				}
				values[k] = v
			}
		}

		// mismatched
		for k, m := range res.Diff.Mismatch {
			fmt.Printf("  %s %s (expected %s, got %q). Enter new value (empty to keep): ", ui.Warn.Render("MISMATCH"), k, m.Expected, m.Actual)
			v, err := ui.PromptValue(k, exampleVals[k], values[k], analyzer.IsSecretKey(k))
			if err != nil {
				if err.Error() == "^C" {
					return err
				}
				logger.Warn("skipping %s: %v", k, err)
				continue
			}
			if v != "" {
				values[k] = v
			}
		}

		// ensure all example keys present in values
		for _, k := range exampleOrder {
			if _, ok := values[k]; !ok {
				values[k] = exampleVals[k]
			}
		}

		envPath := res.EnvFile
		if envPath == "" {
			envPath = filepath.Join(root, ".env")
		}
		if err := config.WriteEnvFile(envPath, exampleOrder, exampleVals, values, extras); err != nil {
			return err
		}
		logger.Info("wrote %s", envPath)
		ui.PrintTitle("Env updated")

		// offer service start
		if res.HasCompose {
			ok, err := ui.Confirm("Start detected background services with docker compose?", true)
			if err == nil && ok {
				if err := startCompose(res.ComposePath); err != nil {
					logger.Warn("%v", err)
				} else {
					ui.Success("services started")
				}
			}
		}
		return nil
	},
}

func mask(s string) string {
	if len(s) <= 6 {
		return "****"
	}
	return s[:3] + "..." + s[len(s)-3:]
}

func startCompose(path string) error {
	return runner.StartServices(path)
}

func init() {
	rootCmd.AddCommand(fixCmd)
}
