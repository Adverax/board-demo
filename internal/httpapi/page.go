package httpapi

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/adverax/board-demo/internal/board"
)

// pageCSP keeps the page self-contained (CON-005): scripts, styles and
// requests only from this origin; the favicon is an inline data: URL.
const pageCSP = "default-src 'self'; img-src 'self' data:; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// NewApp returns the whole application handler: the API of New plus the board
// page (SCR-001) from page, served by the same process (CON-005):
//
//	GET /               page's index.html
//	GET /static/{name}  page's static/<name> (404 when absent)
//
// page is the embedded page of package web (web.Page()). A page without
// index.html is an error, so a broken embed stops the binary at startup.
func NewApp(b *board.Board, page fs.FS) (http.Handler, error) {
	index, err := fs.ReadFile(page, "index.html")
	if err != nil {
		return nil, fmt.Errorf("board page: %w", err)
	}
	static, err := fs.Sub(page, "static")
	if err != nil {
		return nil, fmt.Errorf("board page assets: %w", err)
	}
	mux := apiMux(b)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "no-cache")
		h.Set("Content-Security-Policy", pageCSP)
		_, _ = w.Write(index) // a write error means the client is gone
	})
	mux.HandleFunc("GET /static/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, static, r.PathValue("name"))
	})
	return mux, nil
}
