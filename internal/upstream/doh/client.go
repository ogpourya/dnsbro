package doh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

// Client forwards DNS queries over HTTPS (DoH).
type Client struct {
	Endpoint string
	Client   *http.Client
}

// New creates a DoH client with sane defaults.
func New(endpoint string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
		}).DialContext,
		TLSHandshakeTimeout: timeout,
	}
	return &Client{
		Endpoint: endpoint,
		Client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
	}
}

// Query performs a DoH POST request and returns the DNS response message.
func (c *Client) Query(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	wire, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("pack dns msg: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(wire))
	if err != nil {
		return nil, fmt.Errorf("create doh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform doh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("doh status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read doh body: %w", err)
	}

	var out dns.Msg
	if err := out.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack doh response: %w", err)
	}
	return &out, nil
}
