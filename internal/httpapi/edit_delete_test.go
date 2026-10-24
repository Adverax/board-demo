package httpapi_test

// API/contract tests for PATCH /api/cards/{id} and DELETE /api/cards/{id}
// (CON-003, ADR-0009, CON-004) and for the full CON-003 route set (ADR-0002).

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/adverax/board-demo/internal/board"
	"github.com/adverax/board-demo/internal/httpapi"
)

func patch(t fataler, h http.Handler, id, body string) (int, []byte) {
	t.Helper()
	return do(t, h, http.MethodPatch, "/api/cards/"+id, body)
}

// deleteCard sends DELETE and returns the raw recorder: a 204 carries no body
// and no Content-Type, so the JSON helper do does not fit it.
func deleteCard(h http.Handler, id string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/cards/"+id, nil))
	return rec
}

// cardByID finds a card on the board read through GET /api/board.
func cardByID(t fataler, h http.Handler, id string) cardJSON {
	t.Helper()
	return cardIn(t, getBoard(t, h), id)
}

func cardIn(t fataler, b boardJSON, id string) cardJSON {
	t.Helper()
	for _, col := range b.Columns {
		for _, c := range col.Cards {
			if c.ID == id {
				return c
			}
		}
	}
	t.Fatalf("card %s not on the board", id)
	return cardJSON{}
}

// titled builds a board holding one card titled "Old" in todo.
func titled(t *testing.T) (http.Handler, string) {
	t.Helper()
	b := board.New()
	c, err := b.Add("Old", "keep")
	if err != nil {
		t.Fatal(err)
	}
	return httpapi.New(b), c.ID
}

// assertUnchanged asserts the card still reads as "Old" with description "keep".
func assertUnchanged(t *testing.T, h http.Handler, id string) {
	t.Helper()
	if c := cardByID(t, h, id); c.Title != "Old" || c.Description != "keep" || c.Column != "todo" || c.Position != 0 {
		t.Fatalf("refused edit changed the card: %+v", c)
	}
}

// AC-006
func TestEditCard_UpdatesTitleKeepsPlace(t *testing.T) {
	b, cols := arrange(t, 1, 2, 0)
	h := httpapi.New(b)
	a := cols["doing"][1]
	status, raw := patch(t, h, a, `{"title":"New"}`)
	c := decodeAs[cardJSON](t, raw)
	if status != http.StatusOK || c.ID != a || c.Title != "New" || c.Column != "doing" || c.Position != 1 {
		t.Fatalf("got %d %s", status, raw)
	}
	if got := cardByID(t, h, a); got != c {
		t.Fatalf("board shows %+v, response %+v", got, c)
	}
	if got := ids(getBoard(t, h), "doing"); !slices.Equal(got, cols["doing"]) {
		t.Fatalf("doing = %v, want %v", got, cols["doing"])
	}
}

// AC-007
func TestEditCard_UnknownIdNotFound(t *testing.T) {
	status, raw := patch(t, httpapi.New(board.New()), "missing", `{"title":"New"}`)
	refusal(t, status, raw, http.StatusNotFound, "not_found")
}

// AC-008
func TestEditCard_RejectsEmptyTitle(t *testing.T) {
	h, id := titled(t)
	status, raw := patch(t, h, id, `{"title":""}`)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
	assertUnchanged(t, h, id)
}

// ASM-0001: PATCH is a partial update; an absent or null field stays as it is.
func TestEditCard_PartialUpdate(t *testing.T) {
	h, id := titled(t)
	for _, tc := range []struct {
		body        string
		title, desc string
	}{
		{`{"description":"new desc"}`, "Old", "new desc"},
		{`{"title":"Renamed"}`, "Renamed", "new desc"},
		{`{"title":null,"description":null}`, "Renamed", "new desc"},
		{`{}`, "Renamed", "new desc"},
		{`{"description":""}`, "Renamed", ""},
		{`{"title":"  Both  ","description":"d"}`, "Both", "d"},
	} {
		status, raw := patch(t, h, id, tc.body)
		c := decodeAs[cardJSON](t, raw)
		if status != http.StatusOK || c.Title != tc.title || c.Description != tc.desc || c.Column != "todo" || c.Position != 0 {
			t.Fatalf("PATCH %s: got %d %s, want title %q description %q", tc.body, status, raw, tc.title, tc.desc)
		}
	}
}

// A body that cannot be read as an edit is validation_failed and changes nothing.
func TestEditCard_MalformedBodyRejected(t *testing.T) {
	h, id := titled(t)
	for _, body := range []string{
		`not json`, ``, `{"title":5}`, `{"description":7}`, `{"title":["x"]}`,
		`{"title":"ok"} trailing`, `{"title":"ok","pad":"` + strings.Repeat("a", 1<<20) + `"}`,
	} {
		status, raw := patch(t, h, id, body)
		refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
		assertUnchanged(t, h, id)
	}
}

// AC-021
func TestTitle_WhitespaceOnlyRejectedOnEdit(t *testing.T) {
	h, id := titled(t)
	status, raw := patch(t, h, id, `{"title":"   \t"}`)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
	assertUnchanged(t, h, id)
}

// AC-036
func TestTitle_81CharsRejectedOnEdit(t *testing.T) {
	h, id := titled(t)
	status, raw := patch(t, h, id, `{"title":"`+strings.Repeat("a", 81)+`"}`)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
	assertUnchanged(t, h, id)
}

// AC-037
func TestTitle_80CharsAcceptedOnEdit(t *testing.T) {
	h, id := titled(t)
	title := strings.Repeat("a", 80)
	status, raw := patch(t, h, id, `{"title":"`+title+`"}`)
	if c := decodeAs[cardJSON](t, raw); status != http.StatusOK || c.Title != title {
		t.Fatalf("got %d %s", status, raw)
	}
	if c := cardByID(t, h, id); c.Title != title {
		t.Fatalf("board shows %+v", c)
	}
}

// AC-019
func TestTitle_80CharsAccepted(t *testing.T) {
	h := httpapi.New(board.New())
	title := strings.Repeat("a", 80)
	status, raw := do(t, h, http.MethodPost, "/api/cards", `{"title":"`+title+`"}`)
	c := decodeAs[cardJSON](t, raw)
	if status != http.StatusCreated || c.Title != title {
		t.Fatalf("got %d %s", status, raw)
	}
	if got := ids(getBoard(t, h), "todo"); !slices.Equal(got, []string{c.ID}) {
		t.Fatalf("todo = %v", got)
	}
}

// AC-020
func TestTitle_81CharsRejected(t *testing.T) {
	h := httpapi.New(board.New())
	status, raw := do(t, h, http.MethodPost, "/api/cards", `{"title":"`+strings.Repeat("a", 81)+`"}`)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
	if got := ids(getBoard(t, h), "todo"); len(got) != 0 {
		t.Fatalf("a card was created: %v", got)
	}
}

// ASM-0002: the limit counts runes after trimming, not bytes.
func TestTitle_LimitCountsRunesAfterTrimming(t *testing.T) {
	h := httpapi.New(board.New())
	runes80 := strings.Repeat("я", 80) // 160 bytes
	status, raw := do(t, h, http.MethodPost, "/api/cards", `{"title":"  `+runes80+`\t"}`)
	if c := decodeAs[cardJSON](t, raw); status != http.StatusCreated || c.Title != runes80 {
		t.Fatalf("got %d %s", status, raw)
	}
	status, raw = do(t, h, http.MethodPost, "/api/cards", `{"title":"`+runes80+`я"}`)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
}

// AC-009
func TestDeleteCard_RemovesAndClosesGap(t *testing.T) {
	b, cols := arrange(t, 3, 0, 0)
	h := httpapi.New(b)
	a, bID, c := cols["todo"][0], cols["todo"][1], cols["todo"][2]
	rec := deleteCard(h, bID)
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("DELETE = %d %q, want 204 with no body", rec.Code, rec.Body.String())
	}
	got := getBoard(t, h)
	if !slices.Equal(ids(got, "todo"), []string{a, c}) {
		t.Fatalf("todo = %v", ids(got, "todo"))
	}
	if ca, cc := cardByID(t, h, a), cardByID(t, h, c); ca.Position != 0 || cc.Position != 1 {
		t.Fatalf("positions: A %d, C %d", ca.Position, cc.Position)
	}
}

// AC-010
func TestDeleteCard_UnknownIdNotFound(t *testing.T) {
	b, _ := arrange(t, 2, 1, 1)
	h := httpapi.New(b)
	before := getBoard(t, h)
	rec := deleteCard(h, "missing")
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
	refusal(t, rec.Code, rec.Body.Bytes(), http.StatusNotFound, "not_found")
	after := getBoard(t, h)
	for _, col := range []string{"todo", "doing", "done"} {
		if !slices.Equal(ids(before, col), ids(after, col)) {
			t.Fatalf("refused delete changed %s", col)
		}
	}
}

// CON-003, ADR-0002: the standard ServeMux serves exactly the five documented
// routes; another method on a known path is 405, an unknown path 404.
func TestAPI_ServesTheFiveCON003Routes(t *testing.T) {
	b, cols := arrange(t, 2, 0, 0)
	h := httpapi.New(b)
	id := cols["todo"][0]
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{
		{http.MethodGet, "/api/board", "", http.StatusOK},
		{http.MethodPost, "/api/cards", `{"title":"x"}`, http.StatusCreated},
		{http.MethodPatch, "/api/cards/" + id, `{"title":"y"}`, http.StatusOK},
		{http.MethodPost, "/api/cards/" + id + "/move", `{"to":"doing"}`, http.StatusOK},
		{http.MethodDelete, "/api/cards/" + id, "", http.StatusNoContent},
		{http.MethodPut, "/api/cards/" + cols["todo"][1], `{"title":"z"}`, http.StatusMethodNotAllowed},
		{http.MethodGet, "/api/cards/" + cols["todo"][1], "", http.StatusMethodNotAllowed},
		{http.MethodGet, "/api/archive", "", http.StatusNotFound},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if rec.Code != tc.want {
			t.Fatalf("%s %s = %d %s, want %d", tc.method, tc.path, rec.Code, rec.Body, tc.want)
		}
	}
}

// editBody is one PATCH body variant: what it sets, or unreadable.
type editBody struct {
	raw        string
	readable   bool
	title      *string
	validTitle bool // meaningful when title != nil
}

var editBodies = []editBody{
	{`{"title":"New"}`, true, str("New"), true},
	{`{"title":"  padded  "}`, true, str("padded"), true},
	{`{"title":"` + strings.Repeat("a", 80) + `"}`, true, str(strings.Repeat("a", 80)), true},
	{`{"title":"` + strings.Repeat("a", 81) + `"}`, true, str(""), false},
	{`{"title":""}`, true, str(""), false},
	{`{"title":" \t"}`, true, str(""), false},
	{`{"description":"d"}`, true, nil, false},
	{`{}`, true, nil, false},
	{`{"title":5}`, false, nil, false},
	{`not json`, false, nil, false},
	{``, false, nil, false},
}

// CON-004, INV-001, INV-004 over the API: for any PATCH or DELETE on any
// board, the outcome is the documented one (unreadable body or broken title
// is validation_failed, unknown id is not_found, success returns the card in
// place or 204), a refusal changes nothing, and every column stays dense.
func TestPropertyTest_EditDelete_ContractHolds(t *testing.T) {
	order := []string{"todo", "doing", "done"}
	rapid.Check(t, func(t *rapid.T) {
		b, cols := arrange(t,
			rapid.IntRange(0, 3).Draw(t, "todo"),
			rapid.IntRange(0, board.WIPLimit).Draw(t, "doing"),
			rapid.IntRange(0, 3).Draw(t, "done"))
		h := httpapi.New(b)
		var known []string
		for _, c := range order {
			known = append(known, cols[c]...)
		}
		id := rapid.SampledFrom(append(known, "missing")).Draw(t, "id")
		exists := id != "missing"
		before := getBoard(t, h)

		var changed bool
		if rapid.Bool().Draw(t, "delete") {
			rec := deleteCard(h, id)
			if !exists {
				refusal(t, rec.Code, rec.Body.Bytes(), http.StatusNotFound, "not_found")
			} else if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
				t.Fatalf("DELETE %s = %d %q", id, rec.Code, rec.Body.String())
			}
			changed = exists
		} else {
			body := rapid.SampledFrom(editBodies).Draw(t, "body")
			status, raw := patch(t, h, id, body.raw)
			switch {
			case !body.readable:
				refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
			case !exists:
				refusal(t, status, raw, http.StatusNotFound, "not_found")
			case body.title != nil && !body.validTitle:
				refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
			default:
				old := cardIn(t, before, id)
				c := decodeAs[cardJSON](t, raw)
				if status != http.StatusOK || c.ID != id || c.Column != old.Column || c.Position != old.Position {
					t.Fatalf("PATCH %s = %d %s, card was %+v", body.raw, status, raw, old)
				}
				wantTitle := old.Title
				if body.title != nil {
					wantTitle = *body.title
				}
				if c.Title != wantTitle || cardByID(t, h, id) != c {
					t.Fatalf("PATCH %s returned %+v, want title %q and the board to agree", body.raw, c, wantTitle)
				}
				changed = true
			}
		}

		after := getBoard(t, h)
		for _, col := range after.Columns {
			for pos, c := range col.Cards {
				if c.Position != pos || c.Column != col.Name {
					t.Fatalf("column %s not dense: %+v", col.Name, col.Cards)
				}
			}
		}
		if !changed {
			for i := range before.Columns {
				if !slices.Equal(before.Columns[i].Cards, after.Columns[i].Cards) {
					t.Fatalf("refused request changed %s", before.Columns[i].Name)
				}
			}
		}
	})
}
