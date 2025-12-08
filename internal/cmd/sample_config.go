package cmd

import (
	"fmt"

	"github.com/ogpourya/dnsbro/configs"
	"github.com/spf13/cobra"
)

var sampleConfigCmd = &cobra.Command{
	Use:   "sample-config",
	Short: "Print a sample config to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), configs.SampleYAML)
		return err
	},
}
