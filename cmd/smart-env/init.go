package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/smartenv/smart-env/internal/config"
	"github.com/smartenv/smart-env/internal/ui"
)

var initCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "init",
	Short:        "Initialize .env from example template",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := filepath.Abs(rootPath)
		envPath, err := config.EnsureEnv(root)
		if err != nil {
			return err
		}
		ui.PrintTitle("Initialized")
		fmt.Printf("Created/ensured: %s\n", envPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
