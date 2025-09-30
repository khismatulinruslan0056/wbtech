package httpserver

import (
	"L0/internal/config"
	"L0/internal/service"
	"L0/internal/transport/httpserver/router"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func waitReady(t *testing.T, baseURL, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastErr error
	url := "http://" + baseURL + path
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		lastErr = err
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("server not ready at %s: %v", url, lastErr)
}

func TestHTTPServer_New_ConfigApplied(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	addr := "127.0.0.1:12345"
	cfg := &config.HTTPServer{
		Address:     addr,
		Timeout:     2 * time.Second,
		IdleTimeout: 3 * time.Second,
	}

	var s service.Service
	r := router.New(&s, log)

	srv := New(cfg, r, log)

	if got := srv.Addr(); got != addr {
		t.Fatalf("Addr() = %s, want %s", got, addr)
	}
	if srv.server.ReadTimeout != cfg.Timeout {
		t.Fatalf("ReadTimeout = %s, want %s", srv.server.ReadTimeout, cfg.Timeout)
	}
	if srv.server.WriteTimeout != cfg.Timeout {
		t.Fatalf("WriteTimeout = %s, want %s", srv.server.WriteTimeout, cfg.Timeout)
	}
	if srv.server.IdleTimeout != cfg.IdleTimeout {
		t.Fatalf("IdleTimeout = %s, want %s", srv.server.IdleTimeout, cfg.IdleTimeout)
	}
}

func TestHTTPServer_StartAndShutdown_Graceful(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	addr := freeAddr(t)
	cfg := &config.HTTPServer{
		Address:     addr,
		Timeout:     2 * time.Second,
		IdleTimeout: 2 * time.Second,
	}

	var s service.Service
	r := router.New(&s, log)

	srv := New(cfg, r, log)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	waitReady(t, addr, "/ping", 2*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}

	resp, err := http.Get("http://" + addr + "/ping")
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("server still responds after Shutdown")
	}
}

func TestHTTPServer_Start_FailsOnBusyPort(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pre-listen: %v", err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	cfg := &config.HTTPServer{
		Address:     addr,
		Timeout:     1 * time.Second,
		IdleTimeout: 1 * time.Second,
	}

	var s service.Service
	r := router.New(&s, log)
	srv := New(cfg, r, log)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	select {
	case err = <-errCh:
		if err == nil {
			t.Fatal("expected error due to busy port, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not fail on busy port in time")
	}
}

func TestHTTPServer_HandlesRequests(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	addr := freeAddr(t)
	cfg := &config.HTTPServer{
		Address:     addr,
		Timeout:     time.Second,
		IdleTimeout: time.Second,
	}
	var s service.Service
	r := router.New(&s, log)
	srv := New(cfg, r, log)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		<-errCh
	})

	waitReady(t, addr, "/ping", 2*time.Second)

	resp, err := http.Get("http://" + addr + "/ping")
	if err != nil {
		t.Fatalf("GET /ping error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
