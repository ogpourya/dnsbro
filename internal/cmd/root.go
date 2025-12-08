package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "dnsbro",
	Short: "dnsbro is a local DNS resolver with DoH forwarding and TUI",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "/etc/dnsbro/config.yaml", "Path to config file")
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(reloadCmd)
	rootCmd.AddCommand(revertCmd)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureConfigPath() error {
	if configPath == "" {
		return fmt.Errorf("config path required")
	}
	return nil
}
