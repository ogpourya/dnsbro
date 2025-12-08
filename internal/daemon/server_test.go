package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestQueryWithRetrySucceedsAfterRetries(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	resp, err := queryWithRetry(ctx, 3, 10*time.Millisecond, func(context.Context) (*dns.Msg, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary failure")
		}
		return new(dns.Msg), nil
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}
}

func TestQueryWithRetryHonorsContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := queryWithRetry(ctx, 3, 20*time.Millisecond, func(context.Context) (*dns.Msg, error) {
		return nil, errors.New("fail")
	})
	if err == nil {
		t.Fatalf("expected context error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatalf("context should have cancelled early, took %v", time.Since(start))
	}
}
