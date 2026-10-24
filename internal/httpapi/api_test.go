package httpapi_test

// API/contract tests: every route of this increment against the documented
// request/response shapes (CON-003, ADR-0009) and the CON-004 error codes
// (ADR-0007). They run the real handler on a real Board.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"pgregory.net/rapid"

	"github.com/adverax/board-demo/internal/board"
	"github.com/adverax/board-demo/internal/httpapi"
)

type cardJSON struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Column      string `json:"column"`
	Position    int    `json:"position"`
}

type boardJSON struct {
	Columns []struct {
		Name  string     `json:"name"`
		Cards []cardJSON `json:"cards"`
	} `json:"columns"`
}

type fataler interface {
	Helper()
	Fatalf(format string, args ...any)
}

func do(t fataler, h http.Handler, method, path, body string) (int, []byte) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("%s %s: Content-Type = %q", method, path, ct)
	}
	return rec.Code, rec.Body.Bytes()
}

func decodeAs[T any](t fataler, raw []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return v
}

func getBoard(t fataler, h http.Handler) boardJSON {
	t.Helper()
	status, raw := do(t, h, http.MethodGet, "/api/board", "")
	if status != http.StatusOK {
		t.Fatalf("GET /api/board = %d", status)
	}
	return decodeAs[boardJSON](t, raw)
}

func ids(b boardJSON, column string) []string {
	for _, c := range b.Columns {
		if c.Name == column {
			out := []string{}
			for _, card := range c.Cards {
				out = append(out, card.ID)
			}
			return out
		}
	}
	return nil
}

// refusal asserts a CON-004 error object with the given code and status.
func refusal(t fataler, status int, raw []byte, wantStatus int, wantCode string) {
	t.Helper()
	got := decodeAs[map[string]any](t, raw)
	if status != wantStatus || got["code"] != wantCode || len(got) != 1 {
		t.Fatalf("got %d %s, want %d {\"code\":%q}", status, raw, wantStatus, wantCode)
	}
}

func move(t fataler, h http.Handler, id, to string) (int, []byte) {
	t.Helper()
	return do(t, h, http.MethodPost, "/api/cards/"+id+"/move", `{"to":"`+to+`"}`)
}

// arrange builds a board through the Board API: n cards in todo, then the
// given number of them moved into doing and done. It returns ids per column.
func arrange(t fataler, todo, doing, done int) (*board.Board, map[string][]string) {
	t.Helper()
	b := board.New()
	add := func() string {
		c, err := b.Add("card", "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		return c.ID
	}
	mv := func(id string, to board.Column) {
		if _, err := b.Move(id, &to); err != nil {
			t.Fatalf("Move: %v", err)
		}
	}
	out := map[string][]string{}
	for range done {
		id := add()
		mv(id, board.Doing)
		mv(id, board.Done)
		out["done"] = append(out["done"], id)
	}
	for range doing {
		id := add()
		mv(id, board.Doing)
		out["doing"] = append(out["doing"], id)
	}
	for range todo {
		out["todo"] = append(out["todo"], add())
	}
	return b, out
}

// AC-001, INV-005
func TestGetBoard_GroupsCardsByColumnInOrder(t *testing.T) {
	b, cols := arrange(t, 1, 2, 0)
	got := getBoard(t, httpapi.New(b))
	if len(got.Columns) != 3 || got.Columns[0].Name != "todo" || got.Columns[1].Name != "doing" || got.Columns[2].Name != "done" {
		t.Fatalf("columns = %+v", got.Columns)
	}
	if !slices.Equal(ids(got, "todo"), cols["todo"]) || !slices.Equal(ids(got, "doing"), cols["doing"]) || len(ids(got, "done")) != 0 {
		t.Fatalf("board = %+v, want %v", got, cols)
	}
	for _, c := range got.Columns {
		for pos, card := range c.Cards {
			if card.Position != pos || card.Column != c.Name {
				t.Fatalf("card %+v in %s at %d", card, c.Name, pos)
			}
		}
	}
}

// AC-002
func TestGetBoard_EmptyBoardHasThreeEmptyColumns(t *testing.T) {
	status, raw := do(t, httpapi.New(board.New()), http.MethodGet, "/api/board", "")
	want := `{"columns":[{"name":"todo","cards":[]},{"name":"doing","cards":[]},{"name":"done","cards":[]}]}` + "\n"
	if status != http.StatusOK || string(raw) != want {
		t.Fatalf("got %d %s", status, raw)
	}
}

// AC-003
func TestAddCard_LandsAtEndOfTodo(t *testing.T) {
	b, _ := arrange(t, 2, 0, 0)
	h := httpapi.New(b)
	status, raw := do(t, h, http.MethodPost, "/api/cards", `{"title":"Write docs","description":"draft"}`)
	if status != http.StatusCreated {
		t.Fatalf("status = %d %s", status, raw)
	}
	c := decodeAs[cardJSON](t, raw)
	if c.ID == "" || c.Title != "Write docs" || c.Description != "draft" || c.Column != "todo" || c.Position != 2 {
		t.Fatalf("card = %+v", c)
	}
	if got := ids(getBoard(t, h), "todo"); len(got) != 3 || got[2] != c.ID {
		t.Fatalf("todo = %v", got)
	}
}

// AC-004
func TestAddCard_DescriptionIsOptional(t *testing.T) {
	status, raw := do(t, httpapi.New(board.New()), http.MethodPost, "/api/cards", `{"title":"Write docs"}`)
	c := decodeAs[cardJSON](t, raw)
	if status != http.StatusCreated || c.Description != "" || c.Position != 0 {
		t.Fatalf("got %d %+v", status, c)
	}
}

// AC-005 and malformed add bodies.
func TestAddCard_RejectsBlankTitle(t *testing.T) {
	b, _ := arrange(t, 2, 0, 0)
	h := httpapi.New(b)
	for _, body := range []string{
		`{"title":"   "}`, `{"title":""}`, `{}`, `{"title":5}`, `not json`, ``,
		`{"title":"ok"} trailing`, `{"title":"ok","description":7}`,
	} {
		status, raw := do(t, h, http.MethodPost, "/api/cards", body)
		refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
	}
	if n := len(ids(getBoard(t, h), "todo")); n != 2 {
		t.Fatalf("todo holds %d cards, want 2", n)
	}
}

func TestAddCard_OversizedBodyRejected(t *testing.T) {
	h := httpapi.New(board.New())
	body := `{"title":"x","description":"` + strings.Repeat("a", 1<<20) + `"}`
	status, raw := do(t, h, http.MethodPost, "/api/cards", body)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
}

// AC-011
func TestMoveCard_RightLandsAtEnd(t *testing.T) {
	b, cols := arrange(t, 1, 1, 0)
	h := httpapi.New(b)
	a := cols["todo"][0]
	status, raw := move(t, h, a, "doing")
	c := decodeAs[cardJSON](t, raw)
	if status != http.StatusOK || c.ID != a || c.Column != "doing" || c.Position != 1 {
		t.Fatalf("got %d %+v", status, c)
	}
	got := getBoard(t, h)
	if len(ids(got, "todo")) != 0 || !slices.Equal(ids(got, "doing"), []string{cols["doing"][0], a}) {
		t.Fatalf("board = %+v", got)
	}
}

// AC-012
func TestMoveCard_LeftLandsAtEnd(t *testing.T) {
	b, cols := arrange(t, 2, 1, 0)
	h := httpapi.New(b)
	a := cols["doing"][0]
	status, raw := move(t, h, a, "todo")
	c := decodeAs[cardJSON](t, raw)
	if status != http.StatusOK || c.Column != "todo" || c.Position != 2 {
		t.Fatalf("got %d %+v", status, c)
	}
	if got := ids(getBoard(t, h), "todo"); !slices.Equal(got, append(cols["todo"], a)) {
		t.Fatalf("todo = %v", got)
	}
}

// AC-013
func TestMoveCard_UnknownIdNotFound(t *testing.T) {
	status, raw := move(t, httpapi.New(board.New()), "missing", "doing")
	refusal(t, status, raw, http.StatusNotFound, "not_found")
}

// AC-014
func TestMoveCard_TodoToDoneRefused(t *testing.T) {
	b, cols := arrange(t, 1, 0, 0)
	h := httpapi.New(b)
	status, raw := move(t, h, cols["todo"][0], "done")
	refusal(t, status, raw, http.StatusConflict, "invalid_move")
	if got := ids(getBoard(t, h), "todo"); !slices.Equal(got, cols["todo"]) {
		t.Fatalf("todo = %v", got)
	}
}

// AC-015
func TestMoveCard_DoneToTodoRefused(t *testing.T) {
	b, cols := arrange(t, 0, 0, 1)
	h := httpapi.New(b)
	status, raw := move(t, h, cols["done"][0], "todo")
	refusal(t, status, raw, http.StatusConflict, "invalid_move")
	if got := ids(getBoard(t, h), "done"); !slices.Equal(got, cols["done"]) {
		t.Fatalf("done = %v", got)
	}
}

// ADR-0007/R1: an empty "to" is a present string naming an unknown column, so
// it is invalid_move, not validation_failed (which means "no readable target").
func TestMoveCard_EmptyTargetIsInvalidMove(t *testing.T) {
	b, cols := arrange(t, 1, 0, 0)
	h := httpapi.New(b)
	status, raw := do(t, h, http.MethodPost, "/api/cards/"+cols["todo"][0]+"/move", `{"to":""}`)
	refusal(t, status, raw, http.StatusConflict, "invalid_move")
	if got := ids(getBoard(t, h), "todo"); !slices.Equal(got, cols["todo"]) {
		t.Fatalf("todo = %v", got)
	}
}

// The 1 MiB body cap holds on the move route too: an oversized body is no
// readable target (validation_failed), and not_found still wins (ADR-0007).
func TestMoveCard_OversizedBodyRejected(t *testing.T) {
	b, cols := arrange(t, 1, 0, 0)
	h := httpapi.New(b)
	body := `{"to":"doing","pad":"` + strings.Repeat("a", 1<<20) + `"}`
	status, raw := do(t, h, http.MethodPost, "/api/cards/"+cols["todo"][0]+"/move", body)
	refusal(t, status, raw, http.StatusBadRequest, "validation_failed")
	status, raw = do(t, h, http.MethodPost, "/api/cards/missing/move", body)
	refusal(t, status, raw, http.StatusNotFound, "not_found")
	got := getBoard(t, h)
	if !slices.Equal(ids(got, "todo"), cols["todo"]) || len(ids(got, "doing")) != 0 {
		t.Fatalf("board = %+v", got)
	}
}

// AC-016
func TestMoveCard_IntoDoingAtTwoSucceeds(t *testing.T) {
	b, cols := arrange(t, 1, 2, 0)
	h := httpapi.New(b)
	if status, raw := move(t, h, cols["todo"][0], "doing"); status != http.StatusOK {
		t.Fatalf("got %d %s", status, raw)
	}
	if n := len(ids(getBoard(t, h), "doing")); n != 3 {
		t.Fatalf("doing holds %d", n)
	}
}

// AC-017
func TestMoveCard_IntoFullDoingRefused(t *testing.T) {
	b, cols := arrange(t, 1, 3, 0)
	h := httpapi.New(b)
	status, raw := move(t, h, cols["todo"][0], "doing")
	refusal(t, status, raw, http.StatusConflict, "wip_limit_exceeded")
	got := getBoard(t, h)
	if !slices.Equal(ids(got, "doing"), cols["doing"]) || !slices.Equal(ids(got, "todo"), cols["todo"]) {
		t.Fatalf("board = %+v", got)
	}
}

// AC-033
func TestMoveCard_LeftIntoFullDoingRefused(t *testing.T) {
	b, cols := arrange(t, 0, 3, 1)
	h := httpapi.New(b)
	status, raw := move(t, h, cols["done"][0], "doing")
	refusal(t, status, raw, http.StatusConflict, "wip_limit_exceeded")
	got := getBoard(t, h)
	if !slices.Equal(ids(got, "doing"), cols["doing"]) || !slices.Equal(ids(got, "done"), cols["done"]) {
		t.Fatalf("board = %+v", got)
	}
}

// AC-018: 20 simultaneous HTTP moves into an empty doing over a real server.
func TestMoveCard_20ConcurrentMovesKeepDoingAtMost3(t *testing.T) {
	b, cols := arrange(t, 20, 0, 0)
	srv := httptest.NewServer(httpapi.New(b))
	defer srv.Close()

	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[string]int{}
	for _, id := range cols["todo"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			resp, err := http.Post(srv.URL+"/api/cards/"+id+"/move", "application/json", strings.NewReader(`{"to":"doing"}`))
			if err != nil {
				t.Errorf("POST: %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
			raw, _ := io.ReadAll(resp.Body)
			key := "ok"
			if resp.StatusCode != http.StatusOK {
				var e struct{ Code string }
				_ = json.Unmarshal(raw, &e) // an undecodable body counts under key ""
				key = e.Code
			}
			mu.Lock()
			codes[key]++
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()

	if codes["ok"] != 3 || codes["wip_limit_exceeded"] != 17 || len(codes) != 2 {
		t.Fatalf("outcomes = %v, want 3 ok and 17 wip_limit_exceeded", codes)
	}
	if n := len(b.Snapshot().Lanes[1].Cards); n != 3 {
		t.Fatalf("doing holds %d", n)
	}
}

// moveBody is one request body variant and the target it means (nil when the
// body cannot be read as a move).
type moveBody struct {
	raw string
	to  *string
}

func str(s string) *string { return &s }

var moveBodies = []moveBody{
	{`{"to":"todo"}`, str("todo")},
	{`{"to":"doing"}`, str("doing")},
	{`{"to":"done"}`, str("done")},
	{`{"to":"archive"}`, str("archive")},
	{`{"to":"DOING"}`, str("DOING")},
	{`{"to":""}`, str("")},
	{`{"to":5}`, nil},
	{`{"to":null}`, nil},
	{`{"to":["doing"]}`, nil},
	{`{}`, nil},
	{`{"from":"todo"}`, nil},
	{`not json`, nil},
	{``, nil},
	{`{"to":"doing"} trailing`, nil},
}

// EC(ADR-0007/R1), ASM-0003: for any move request the refusal reports the
// first failing check in the order not_found, validation_failed, invalid_move,
// wip_limit_exceeded; same-column and unknown column names are invalid_move;
// a refused move changes nothing.
func TestPropertyTest_MoveRefusal_FirstFailingCheckWins(t *testing.T) {
	order := []string{"todo", "doing", "done"}
	rapid.Check(t, func(t *rapid.T) {
		b, cols := arrange(t,
			rapid.IntRange(0, 2).Draw(t, "todo"),
			rapid.IntRange(0, board.WIPLimit).Draw(t, "doing"),
			rapid.IntRange(0, 2).Draw(t, "done"))
		h := httpapi.New(b)
		var known []string
		for _, c := range order {
			known = append(known, cols[c]...)
		}
		id := rapid.SampledFrom(append(known, "missing")).Draw(t, "id")
		body := rapid.SampledFrom(moveBodies).Draw(t, "body")
		before := getBoard(t, h)

		from := -1
		for i, c := range order {
			if slices.Contains(cols[c], id) {
				from = i
			}
		}
		wantStatus, wantCode := http.StatusOK, ""
		switch {
		case from < 0:
			wantStatus, wantCode = http.StatusNotFound, "not_found"
		case body.to == nil:
			wantStatus, wantCode = http.StatusBadRequest, "validation_failed"
		case !slices.Contains(order, *body.to), slices.Index(order, *body.to) != from-1 && slices.Index(order, *body.to) != from+1:
			wantStatus, wantCode = http.StatusConflict, "invalid_move"
		case *body.to == "doing" && len(cols["doing"]) >= board.WIPLimit:
			wantStatus, wantCode = http.StatusConflict, "wip_limit_exceeded"
		}

		status, raw := do(t, h, http.MethodPost, "/api/cards/"+id+"/move", body.raw)
		if wantCode == "" {
			c := decodeAs[cardJSON](t, raw)
			if status != http.StatusOK || c.ID != id || c.Column != *body.to {
				t.Fatalf("got %d %s, want the moved card", status, raw)
			}
			return
		}
		refusal(t, status, raw, wantStatus, wantCode)
		after := getBoard(t, h)
		for _, c := range order {
			if !slices.Equal(ids(before, c), ids(after, c)) {
				t.Fatalf("refused move changed %s", c)
			}
		}
	})
}
