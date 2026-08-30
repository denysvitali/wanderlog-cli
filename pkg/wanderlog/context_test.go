package wanderlog

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type blockingRoundTripper struct {
	started chan struct{}
}

func (t *blockingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	close(t.started)
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func TestGetTripContextCancelsRequest(t *testing.T) {
	client := NewClient()
	transport := &blockingRoundTripper{started: make(chan struct{})}
	client.httpClient.Transport = transport

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := client.GetTripContext(ctx, "trip-key")
		done <- err
	}()

	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled), "expected context cancellation, got %v", err)
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}

func TestGetUserTripsContextCancelsRequest(t *testing.T) {
	client := NewClient()
	transport := &blockingRoundTripper{started: make(chan struct{})}
	client.httpClient.Transport = transport

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := client.GetUserTripsContext(ctx)
		done <- err
	}()

	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled), "expected context cancellation, got %v", err)
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}

func TestDoAPIContextCancelsRequest(t *testing.T) {
	client := NewClient()
	transport := &blockingRoundTripper{started: make(chan struct{})}
	client.httpClient.Transport = transport

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, _, err := client.DoAPIContext(ctx, http.MethodGet, "health", nil, nil, false)
		done <- err
	}()

	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled), "expected context cancellation, got %v", err)
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}

func TestLoginContextCancelsRequest(t *testing.T) {
	client := NewClient()
	transport := &blockingRoundTripper{started: make(chan struct{})}
	client.httpClient.Transport = transport

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := client.LoginContext(ctx, "traveler@example.com", "secret")
		done <- err
	}()

	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled), "expected context cancellation, got %v", err)
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}

func TestDoAPIContextRejectsNilContext(t *testing.T) {
	client := NewClient()
	//nolint:staticcheck // Explicitly verifies the defensive nil-context contract.
	_, _, err := client.DoAPIContext(nil, http.MethodGet, "health", nil, nil, false)
	require.EqualError(t, err, "creating request: nil context")
}
