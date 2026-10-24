// Package httpapi is the JSON API of the board (COMP-002): the CON-003 routes
// on the standard net/http ServeMux (ADR-0002), request decoding and the
// CON-004 error object. Handlers reach cards only through the exported Board
// API (ADR-0004/R2); every rule lives in package board.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/adverax/board-demo/internal/board"
)

// maxBodyBytes caps every request body; a larger body is validation_failed.
const maxBodyBytes = 1 << 20

// The stable error codes of CON-004.
const (
	codeValidationFailed = "validation_failed"
	codeNotFound         = "not_found"
	codeWIPLimitExceeded = "wip_limit_exceeded"
	codeInvalidMove      = "invalid_move"
)

type cardJSON struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Column      string `json:"column"`
	Position    int    `json:"position"`
}

type columnJSON struct {
	Name  string     `json:"name"`
	Cards []cardJSON `json:"cards"`
}

type boardJSON struct {
	Columns []columnJSON `json:"columns"`
}

type errorJSON struct {
	Code string `json:"code"`
}

// New returns the API handler serving the five CON-003 routes on b.
func New(b *board.Board) http.Handler {
	a := &api{board: b}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/board", a.getBoard)
	mux.HandleFunc("POST /api/cards", a.addCard)
	mux.HandleFunc("PATCH /api/cards/{id}", a.editCard)
	mux.HandleFunc("POST /api/cards/{id}/move", a.moveCard)
	mux.HandleFunc("DELETE /api/cards/{id}", a.deleteCard)
	return mux
}

type api struct {
	board *board.Board
}

func (a *api) getBoard(w http.ResponseWriter, _ *http.Request) {
	s := a.board.Snapshot()
	out := boardJSON{Columns: make([]columnJSON, 0, len(s.Lanes))}
	for _, l := range s.Lanes {
		cards := make([]cardJSON, 0, len(l.Cards))
		for _, c := range l.Cards {
			cards = append(cards, toJSON(c))
		}
		out.Columns = append(out.Columns, columnJSON{Name: string(l.Name), Cards: cards})
	}
	writeJSON(w, http.StatusOK, out)
}

// addCard: body {"title": string, "description"?: string}; 201 with the card.
func (a *api) addCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := decode(w, r, &req); err != nil {
		writeError(w, board.ErrValidation)
		return
	}
	c, err := a.board.Add(req.Title, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toJSON(c))
}

// editCard: body {"title"?: string, "description"?: string}; 200 with the
// edited card (ADR-0009). A partial update (ASM-0001): an absent or null field
// stays as it is. A body that cannot be read as an edit (malformed, oversized,
// a non-string field) is validation_failed; an unknown id is not_found; a
// title breaking INV-001 is validation_failed and changes nothing.
func (a *api) editCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
	}
	if err := decode(w, r, &req); err != nil {
		writeError(w, board.ErrValidation)
		return
	}
	c, err := a.board.Edit(r.PathValue("id"), req.Title, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toJSON(c))
}

// deleteCard: no body; 204 with no body (ADR-0009), or not_found.
func (a *api) deleteCard(w http.ResponseWriter, r *http.Request) {
	if err := a.board.Delete(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// moveCard: body {"to": "<column>"}; 200 with the moved card. A body that
// cannot be read as a move (malformed, oversized, missing or non-string "to")
// passes a nil target, so the aggregate reports not_found before
// validation_failed in one critical section (ADR-0007). Any string "to", the
// empty string included, is a named target: unknown names are invalid_move.
func (a *api) moveCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To *string `json:"to"`
	}
	var to *board.Column
	if err := decode(w, r, &req); err == nil && req.To != nil {
		c := board.Column(*req.To)
		to = &c
	}
	c, err := a.board.Move(r.PathValue("id"), to)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toJSON(c))
}

// decode reads at most maxBodyBytes and unmarshals exactly one JSON value.
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func toJSON(c board.Card) cardJSON {
	return cardJSON{ID: c.ID, Title: c.Title, Description: c.Description, Column: string(c.Column), Position: c.Position}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // a write error means the client is gone
}

// writeError is the one JSON error helper: it maps a board refusal onto its
// CON-004 code and HTTP status.
func writeError(w http.ResponseWriter, err error) {
	var status int
	var code string
	switch {
	case errors.Is(err, board.ErrNotFound):
		status, code = http.StatusNotFound, codeNotFound
	case errors.Is(err, board.ErrValidation):
		status, code = http.StatusBadRequest, codeValidationFailed
	case errors.Is(err, board.ErrInvalidMove):
		status, code = http.StatusConflict, codeInvalidMove
	case errors.Is(err, board.ErrWIPLimit):
		status, code = http.StatusConflict, codeWIPLimitExceeded
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, status, errorJSON{Code: code})
}
