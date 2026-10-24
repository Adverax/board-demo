package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/adverax/board-demo/internal/board"
	"github.com/adverax/board-demo/internal/httpapi"
)

var testPage = fstest.MapFS{
	"index.html":    {Data: []byte("<!doctype html><title>Team board</title>")},
	"static/app.js": {Data: []byte("'use strict';")},
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// GET / serves index.html, GET /static/{name} its assets, anything else is 404;
// the API routes are the same as New's.
func TestNewApp_ServesPageAndAssets(t *testing.T) {
	h, err := httpapi.NewApp(board.New(), testPage)
	if err != nil {
		t.Fatal(err)
	}
	rec := get(t, h, "/")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<title>Team board</title>") {
		t.Fatalf("GET / = %d %q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("GET / Content-Type = %q", ct)
	}
	if csp := rec.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
		t.Fatalf("GET / CSP = %q", csp)
	}
	rec = get(t, h, "/static/app.js")
	if rec.Code != http.StatusOK || rec.Body.String() != "'use strict';" || !strings.Contains(rec.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("GET /static/app.js = %d %q %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
	for _, path := range []string{"/static/missing.css", "/index.html", "/other"} {
		if rec := get(t, h, path); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
	}
	rec = get(t, h, "/api/board")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("GET /api/board = %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / = %d, want 405", rec.Code)
	}
}

// A page without index.html cannot be served, so NewApp refuses it and the
// binary stops before it listens.
func TestNewApp_MissingIndexIsAnError(t *testing.T) {
	if _, err := httpapi.NewApp(board.New(), fstest.MapFS{"static/app.js": {}}); err == nil {
		t.Fatal("NewApp without index.html returned no error")
	}
}
