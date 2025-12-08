package daemon

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ogpourya/dnsbro/internal/logging"
	"github.com/ogpourya/dnsbro/internal/rules"
	"github.com/ogpourya/dnsbro/internal/upstream/doh"
	"github.com/ogpourya/dnsbro/pkg/config"

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
	cfg    config.Config
	rules  rules.RuleSet
	logger *logging.Logger
	doh    *doh.Client
	mu     sync.RWMutex
	stats  Stats
}

// New returns a configured Daemon.
func New(cfg config.Config, logger *logging.Logger) *Daemon {
	r := rules.RuleSet{
		Blocklist: cfg.Rules.Blocklist,
		Allowlist: cfg.Rules.Allowlist,
	}
	return &Daemon{
		cfg:    cfg,
		rules:  r,
		logger: logger,
		doh:    doh.New(cfg.Upstream.DoHEndpoint, cfg.Upstream.Timeout, cfg.Upstream.Bootstrap),
	}
}

// Reload swaps the daemon configuration at runtime.
func (d *Daemon) Reload(cfg config.Config) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cfg = cfg
	d.rules = rules.RuleSet{
		Blocklist: cfg.Rules.Blocklist,
		Allowlist: cfg.Rules.Allowlist,
	}
	d.doh = doh.New(cfg.Upstream.DoHEndpoint, cfg.Upstream.Timeout, cfg.Upstream.Bootstrap)
	d.logger.Infof("configuration reloaded")
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

	d.logger.Infof("dnsbro listening on %s (udp/tcp)", d.cfg.Listen)

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

	d.mu.RLock()
	cfg := d.cfg
	rs := d.rules
	upstream := d.doh
	d.mu.RUnlock()

	question := r.Question[0]
	domain := question.Name
	clientIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	start := time.Now()
	ev := QueryEvent{
		Domain:   domain,
		Client:   clientIP,
		Upstream: cfg.Upstream.DoHEndpoint,
	}

	if rs.ShouldBlock(domain) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(m)
		ev.Blocked = true
		ev.RCode = m.Rcode
		d.recordEvent(ev)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Upstream.Timeout)
	defer cancel()

	resp, err := queryWithRetry(ctx, 3, time.Second, func(ctx context.Context) (*dns.Msg, error) {
		return upstream.Query(ctx, r)
	})
	if err != nil {
		ev.Err = err
		d.logger.Errorf("doh query failed for %s after retries: %v", domain, err)
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
		d.logger.Infof("blocked %s from %s", ev.Domain, ev.Client)
	} else if ev.Err != nil {
		d.logger.Errorf("error handling %s: %v", ev.Domain, ev.Err)
	} else {
		d.logger.Debugf("resolved %s via %s -> %v", ev.Domain, ev.Upstream, ev.ResponseIPs)
	}

	// Non-blocking send so DNS path isn't stalled by slow UI.
}

func queryWithRetry(ctx context.Context, attempts int, delay time.Duration, fn func(context.Context) (*dns.Msg, error)) (*dns.Msg, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 1; i <= attempts; i++ {
		resp, err := fn(ctx)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if i == attempts {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, lastErr
}
