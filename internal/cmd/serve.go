package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dnsbro/internal/daemon"
	"dnsbro/internal/logging"
	appTUI "dnsbro/internal/tui"
	"dnsbro/pkg/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	noTUI bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the dnsbro DNS daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureConfigPath(); err != nil {
			return err
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				cfg = config.Defaults()
			} else {
				return fmt.Errorf("load config: %w", err)
			}
		}

		logr, err := logging.New(cfg.Log.File, "dnsbro ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: using stdout logging: %v\n", err)
			logr, _ = logging.New("", "dnsbro ")
		}
		defer logr.Close()

		events := make(chan daemon.QueryEvent, 64)
		d := daemon.New(cfg, logr, events)

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		reloadCh := make(chan os.Signal, 1)
		signal.Notify(reloadCh, syscall.SIGHUP)
		go func() {
			for range reloadCh {
				newCfg, err := config.Load(configPath)
				if err != nil {
					logr.Printf("reload failed: %v", err)
					continue
				}
				d.Reload(newCfg)
			}
		}()

		if !noTUI && cfg.TUI.Enabled {
			prog := tea.NewProgram(appTUI.NewModel(events), tea.WithAltScreen())
			go func() {
				if err := prog.Start(); err != nil {
					logr.Printf("tui exited: %v", err)
				}
				stop()
			}()
		} else {
			// Drain events to avoid blocking when TUI disabled.
			go func() {
				for range events {
				}
			}()
		}

		return d.Start(ctx)
	},
}

func init() {
	serveCmd.Flags().BoolVar(&noTUI, "no-tui", false, "Disable TUI even if enabled in config")
}
