package cmd

import (
	"fmt"

	"github.com/ogpourya/dnsbro/internal/daemon"
	appTUI "github.com/ogpourya/dnsbro/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the dnsbro TUI without starting the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Standalone TUI mode does not have live data but helps validate rendering.
		events := make(chan daemon.QueryEvent)
		prog := tea.NewProgram(appTUI.NewModel(events))
		close(events)
		if err := prog.Start(); err != nil {
			return fmt.Errorf("start tui: %w", err)
		}
		return nil
	},
}
