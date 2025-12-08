package cmd

import (
	"bufio"
	"net"
	"os"
	"strings"

	"github.com/ogpourya/dnsbro/internal/logging"
)

// warnIfSystemResolverBypasses logs a warning when the system resolver
// is not pointed at the address dnsbro listens on (common when systemd-resolved
// keeps using 127.0.0.53).
func warnIfSystemResolverBypasses(logr *logging.Logger, listenAddr string) {
	host, _, err := net.SplitHostPort(listenAddr)
	if err != nil {
		logr.Warnf("could not parse listen address %q: %v", listenAddr, err)
		return
	}

	nameservers, err := readNameservers("/etc/resolv.conf")
	if err != nil {
		logr.Warnf("could not read /etc/resolv.conf: %v", err)
		return
	}

	for _, ns := range nameservers {
		if ns == host {
			return
		}
	}

	if len(nameservers) == 0 {
		logr.Warnf("/etc/resolv.conf lists no nameservers; point it to %s so dnsbro handles queries", host)
		return
	}

	logr.Warnf("system nameservers %v do not include %s; DNS queries may bypass dnsbro (e.g. via systemd-resolved). Point your resolver to %s.", nameservers, host, host)
}

func readNameservers(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var nameservers []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.EqualFold(fields[0], "nameserver") {
			nameservers = append(nameservers, fields[1])
		}
	}
	return nameservers, sc.Err()
}
