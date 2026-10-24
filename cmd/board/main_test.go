package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/adverax/board-demo/internal/board"
	"github.com/adverax/board-demo/internal/httpapi"
)

// CON-006: the port comes from PORT, default 8080.
func TestAddr_PortFromEnvDefault8080(t *testing.T) {
	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "PORT" {
				return v
			}
			return ""
		}
	}
	if got := addr(env("")); got != ":8080" {
		t.Fatalf("addr without PORT = %q", got)
	}
	if got := addr(env("9123")); got != ":9123" {
		t.Fatalf("addr with PORT=9123 = %q", got)
	}
}

// The wired server answers the three routes and shuts down within
// shutdownTimeout once its context is cancelled.
func TestServe_AnswersRoutesAndShutsDownOnCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, ln, httpapi.New(board.New())) }()
	base := "http://" + ln.Addr().String()

	post := func(path, body string) (int, string) {
		resp, err := http.Post(base+path, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(raw)
	}
	if status, raw := post("/api/cards", `{"title":"Write docs"}`); status != http.StatusCreated || !strings.Contains(raw, `"id":"1"`) {
		t.Fatalf("add = %d %s", status, raw)
	}
	if status, raw := post("/api/cards/1/move", `{"to":"done"}`); status != http.StatusConflict || !strings.Contains(raw, `"code":"invalid_move"`) {
		t.Fatalf("move = %d %s", status, raw)
	}
	resp, err := http.Get(base + "/api/board")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/board = %v, %v", resp, err)
	}
	resp.Body.Close()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve = %v", err)
		}
	case <-time.After(shutdownTimeout + time.Second):
		t.Fatal("serve did not return within shutdownTimeout")
	}
}
