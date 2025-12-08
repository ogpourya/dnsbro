package cmd

import (
	"fmt"

	"github.com/ogpourya/dnsbro/internal/systemd"
	"github.com/ogpourya/dnsbro/pkg/config"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dnsbro as a systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := ensureConfigPath(); err != nil {
			return err
		}

		if !fileExists(configPath) {
			cfg := config.Defaults()
			if err := config.Write(configPath, cfg); err != nil {
				return err
			}
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if err := systemd.Install(configPath, cfg.Listen); err != nil {
			return fmt.Errorf("install systemd service: %w", err)
		}
		fmt.Printf("dnsbro installed: service=%s binary=%s config=%s\n", systemd.ServicePath(), systemd.BinaryPath(), configPath)
		return nil
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the systemd service and binary",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := systemd.Uninstall(); err != nil {
			return fmt.Errorf("uninstall: %w", err)
		}
		fmt.Println("dnsbro uninstalled; restored /etc/resolv.conf if backup was present")
		return nil
	},
}

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Stop dnsbro and revert DNS settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := systemd.Revert(); err != nil {
			return fmt.Errorf("revert: %w", err)
		}
		fmt.Println("dnsbro stopped; /etc/resolv.conf restored if backup was present")
		return nil
	},
}
