package cmd

import (
	"fmt"
	"os"
)

// requireRoot enforces root execution for commands that need privileged access.
func requireRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("root required: re-run with sudo")
	}
	return nil
}
