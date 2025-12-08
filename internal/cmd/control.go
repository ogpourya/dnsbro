package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the dnsbro systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := exec.Command("systemctl", "start", "dnsbro").Run(); err != nil {
			return fmt.Errorf("systemctl start: %w", err)
		}
		fmt.Println("dnsbro service started")
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the dnsbro systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := exec.Command("systemctl", "stop", "dnsbro").Run(); err != nil {
			return fmt.Errorf("systemctl stop: %w", err)
		}
		fmt.Println("dnsbro service stopped")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the dnsbro systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		out, err := exec.Command("systemctl", "status", "dnsbro").CombinedOutput()
		fmt.Print(string(out))
		return err
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the dnsbro configuration via systemd",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := exec.Command("systemctl", "reload-or-restart", "dnsbro").Run(); err != nil {
			return fmt.Errorf("systemctl reload-or-restart: %w", err)
		}
		fmt.Println("dnsbro service reloaded")
		return nil
	},
}
