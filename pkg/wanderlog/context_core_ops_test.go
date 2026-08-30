package wanderlog

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestApplyOperationsContextCancelsRequest(t *testing.T) {
	client := NewClient()
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "token"})
	transport := &blockingRoundTripper{started: make(chan struct{})}
	client.httpClient.Transport = transport

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- client.ApplyOperationsContext(ctx, "trip-key", nil)
	}()

	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ApplyOperationsContext error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}

func TestSetTripBudgetContextCancelsInitialRead(t *testing.T) {
	client := NewClient()
	transport := &blockingRoundTripper{started: make(chan struct{})}
	client.httpClient.Transport = transport

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- client.SetTripBudgetContext(ctx, "trip-key", 100, "USD")
	}()

	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("SetTripBudgetContext error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}
