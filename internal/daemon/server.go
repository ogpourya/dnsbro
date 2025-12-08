package daemon

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"dnsbro/internal/logging"
	"dnsbro/internal/rules"
	"dnsbro/internal/upstream/doh"
	"dnsbro/pkg/config"

	"github.com/miekg/dns"
)

// QueryEvent represents a single DNS query handled by the daemon.
type QueryEvent struct {
	Domain      string
	Client      string
	Upstream    string
	ResponseIPs []string
	RCode       int
	Duration    time.Duration
	Blocked     bool
	Err         error
}

// Stats stores runtime counters.
type Stats struct {
	mu       sync.Mutex
	Queries  int
	Blocked  int
	Failures int
	Last     QueryEvent
}

// Daemon runs the DNS server and forwards requests to DoH.
type Daemon struct {
	cfg     config.Config
	rules   rules.RuleSet
	logger  *logging.Logger
	doh     *doh.Client
	stats   Stats
	updates chan QueryEvent
}

// New returns a configured Daemon.
func New(cfg config.Config, logger *logging.Logger, updates chan QueryEvent) *Daemon {
	r := rules.RuleSet{
		Blocklist: cfg.Rules.Blocklist,
		Allowlist: cfg.Rules.Allowlist,
	}
	return &Daemon{
		cfg:     cfg,
		rules:   r,
		logger:  logger,
		doh:     doh.New(cfg.Upstream.DoHEndpoint, cfg.Upstream.Timeout),
		updates: updates,
	}
}

// Reload swaps the daemon configuration at runtime.
func (d *Daemon) Reload(cfg config.Config) {
	d.cfg = cfg
	d.rules = rules.RuleSet{
		Blocklist: cfg.Rules.Blocklist,
		Allowlist: cfg.Rules.Allowlist,
	}
	d.doh = doh.New(cfg.Upstream.DoHEndpoint, cfg.Upstream.Timeout)
	d.logger.Printf("configuration reloaded")
}

// Start launches UDP and TCP listeners. Caller should cancel the context to stop.
func (d *Daemon) Start(ctx context.Context) error {
	if d.cfg.Listen == "" {
		return errors.New("listen address missing")
	}

	udpServer := &dns.Server{Addr: d.cfg.Listen, Net: "udp", Handler: d}
	tcpServer := &dns.Server{Addr: d.cfg.Listen, Net: "tcp", Handler: d}

	errCh := make(chan error, 2)

	go func() { errCh <- udpServer.ListenAndServe() }()
	go func() { errCh <- tcpServer.ListenAndServe() }()

	d.logger.Printf("dnsbro listening on %s (udp/tcp)", d.cfg.Listen)

	select {
	case <-ctx.Done():
		_ = udpServer.Shutdown()
		_ = tcpServer.Shutdown()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// ServeDNS implements dns.Handler.
func (d *Daemon) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeFormatError))
		return
	}

	question := r.Question[0]
	domain := question.Name
	clientIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	start := time.Now()
	ev := QueryEvent{
		Domain:   domain,
		Client:   clientIP,
		Upstream: d.cfg.Upstream.DoHEndpoint,
	}

	if d.rules.ShouldBlock(domain) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(m)
		ev.Blocked = true
		ev.RCode = m.Rcode
		d.recordEvent(ev)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.Upstream.Timeout)
	defer cancel()

	resp, err := d.doh.Query(ctx, r)
	if err != nil {
		ev.Err = err
		d.logger.Printf("doh query failed for %s: %v", domain, err)
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeServerFailure
		_ = w.WriteMsg(m)
		ev.RCode = m.Rcode
		d.recordEvent(ev)
		return
	}

	ev.Duration = time.Since(start)
	ev.RCode = resp.Rcode
	for _, ans := range resp.Answer {
		if arec, ok := ans.(*dns.A); ok {
			ev.ResponseIPs = append(ev.ResponseIPs, arec.A.String())
		}
		if aaaa, ok := ans.(*dns.AAAA); ok {
			ev.ResponseIPs = append(ev.ResponseIPs, aaaa.AAAA.String())
		}
	}

	_ = w.WriteMsg(resp)
	d.recordEvent(ev)
}

func (d *Daemon) recordEvent(ev QueryEvent) {
	d.stats.mu.Lock()
	d.stats.Queries++
	if ev.Blocked {
		d.stats.Blocked++
	}
	if ev.Err != nil {
		d.stats.Failures++
	}
	d.stats.Last = ev
	d.stats.mu.Unlock()

	if ev.Blocked {
		d.logger.Printf("blocked %s from %s", ev.Domain, ev.Client)
	} else if ev.Err != nil {
		d.logger.Printf("error handling %s: %v", ev.Domain, ev.Err)
	} else {
		d.logger.Printf("resolved %s via %s -> %v", ev.Domain, ev.Upstream, ev.ResponseIPs)
	}

	// Non-blocking send so DNS path isn't stalled by slow UI.
	select {
	case d.updates <- ev:
	default:
	}
}
