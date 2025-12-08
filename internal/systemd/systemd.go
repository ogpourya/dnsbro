package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	servicePath = "/etc/systemd/system/dnsbro.service"
	binaryPath  = "/usr/local/bin/dnsbro"
)

// Install writes the service file and enables dnsbro via systemctl.
func Install(configPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	if err := copyFile(exe, binaryPath); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}

	unit := serviceUnit(configPath)
	if err := os.WriteFile(servicePath, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("write service: %w", err)
	}

	_ = exec.Command("systemctl", "daemon-reload").Run()
	_ = exec.Command("systemctl", "enable", "--now", "dnsbro").Run()
	return nil
}

// Uninstall removes the service and binary.
func Uninstall() error {
	_ = exec.Command("systemctl", "stop", "dnsbro").Run()
	_ = exec.Command("systemctl", "disable", "dnsbro").Run()
	_ = os.Remove(servicePath)
	return os.Remove(binaryPath)
}

// Revert stops the service and leaves a hook to restore DNS.
func Revert() error {
	_ = exec.Command("systemctl", "stop", "dnsbro").Run()
	// Restoring /etc/resolv.conf is system dependent; we document manual steps.
	return nil
}

func ServicePath() string { return servicePath }
func BinaryPath() string  { return binaryPath }

func serviceUnit(configPath string) string {
	return fmt.Sprintf(`[Unit]
Description=dnsbro local DNS resolver
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=%s serve --config %s --no-tui
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
User=root
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
`, binaryPath, configPath)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o755)
}
