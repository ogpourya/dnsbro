package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ogpourya/dnsbro/internal/daemon"
	"github.com/ogpourya/dnsbro/internal/logging"
	"github.com/ogpourya/dnsbro/pkg/config"

	"github.com/spf13/cobra"
)

var (
	listenOverride string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the dnsbro DNS daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRoot(); err != nil {
			return err
		}

		if err := ensureConfigPath(); err != nil {
			return err
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				cfg = config.Defaults()
				if err := config.Write(configPath, cfg); err != nil {
					return fmt.Errorf("write default config: %w", err)
				}
				fmt.Fprintf(os.Stdout, "created default config at %s\n", configPath)
			} else {
				return fmt.Errorf("load config: %w", err)
			}
		}

		if listenOverride != "" {
			cfg.Listen = listenOverride
		}

		logr, err := logging.New(cfg.Log.File, "dnsbro ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: using stdout logging: %v\n", err)
			logr, _ = logging.New("", "dnsbro ")
		}
		defer logr.Close()

		warnIfSystemResolverBypasses(logr, cfg.Listen)

		d := daemon.New(cfg, logr)

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
				if listenOverride != "" {
					newCfg.Listen = listenOverride
				}
				d.Reload(newCfg)
			}
		}()

		return d.Start(ctx)
	},
}

func init() {
	serveCmd.Flags().StringVar(&listenOverride, "listen", "", "Override listen address (host:port) without editing the config file")
}
