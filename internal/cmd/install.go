package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ogpourya/dnsbro/internal/systemd"
	"github.com/ogpourya/dnsbro/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dnsbro as a systemd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureConfigPath(); err != nil {
			return err
		}

		if !fileExists(configPath) {
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				return err
			}
			cfg := config.Defaults()
			data, err := yamlMarshal(cfg)
			if err != nil {
				return err
			}
			if err := os.WriteFile(configPath, data, 0o644); err != nil {
				return err
			}
		}

		if err := systemd.Install(configPath); err != nil {
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
		if err := systemd.Uninstall(); err != nil {
			return fmt.Errorf("uninstall: %w", err)
		}
		fmt.Println("dnsbro uninstalled")
		return nil
	},
}

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Stop dnsbro and revert DNS settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := systemd.Revert(); err != nil {
			return fmt.Errorf("revert: %w", err)
		}
		fmt.Println("dnsbro stopped; restore /etc/resolv.conf if needed")
		return nil
	},
}

func yamlMarshal(cfg config.Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
