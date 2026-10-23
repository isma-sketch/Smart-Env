package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/smartenv/smart-env/internal/logger"
	"github.com/smartenv/smart-env/internal/ui"
)

var (
	rootPath string
	verbose  bool
	version  = "dev"
)

var rootCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: false,
	Version:       version,
	Use:           "smart-env",
	Short:         "Smart-Env: analyze, check and spin up project environments",
	Long:          "Smart-Env automatically analyzes a repository, detects missing env vars and dependencies, and helps spin up the environment.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// default: scan + check summary
		ui.PrintTitle("Smart-Env")
		fmt.Println(ui.Dim.Render("Run 'smart-env check' for details or 'smart-env fix' to resolve."))
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootPath, "path", "p", ".", "project root path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		logger.SetVerbose(verbose)
		viper.Set("path", rootPath)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
