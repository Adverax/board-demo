package main

// Browser end-to-end tests (ADR-0010, NFR-003): headless Chrome driven through
// chromedp. Run them with `make test-e2e`; `make test` skips TestE2E_*.
//
// Browser tests run the real wiring (httpapi.NewApp over a real Board and the
// embedded web.Page()) on an httptest server bound to a free port, so a test
// can seed the board and act as "another member" through the Board API. The
// two PORT tests build and start the real binary on its fixed port.
//
// No Chrome is a failure, never a skip: a green `make test-e2e` always means
// the page ran in a browser. Every browser test fails on any console error or
// warning, uncaught exception, failed request or unexpected 4xx/5xx response.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	cdplog "github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/adverax/board-demo/internal/board"
	"github.com/adverax/board-demo/internal/httpapi"
	"github.com/adverax/board-demo/web"
)

const (
	e2eTimeout     = 30 * time.Second // whole browser test, Chrome start included
	startTimeout   = 10 * time.Second // the built binary answers GET / within this
	requestTimeout = 5 * time.Second  // REQUEST_TIMEOUT_MS of web/static/app.js
)

// chromeCandidates are tried in order after CHROME_PATH.
var chromeCandidates = []string{
	"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	"/Applications/Chromium.app/Contents/MacOS/Chromium",
	"google-chrome", "google-chrome-stable", "chromium", "chromium-browser",
	"headless-shell", "headless_shell", "chrome",
}

func findChrome() (string, error) {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		found, err := exec.LookPath(p)
		if err != nil {
			return "", fmt.Errorf("CHROME_PATH=%q: %w", p, err)
		}
		return found, nil
	}
	for _, c := range chromeCandidates {
		if found, err := exec.LookPath(c); err == nil {
			return found, nil
		}
	}
	return "", errors.New("e2e: no Chrome/Chromium found (tried CHROME_PATH and " +
		strings.Join(chromeCandidates, ", ") + "); install one or set CHROME_PATH — the browser suite cannot run without it (ADR-0010)")
}

// --- the app under test ------------------------------------------------------

// startApp serves the real application on a free port. wrap, when not nil,
// sits in front of the handler (for slow or stubbed responses).
func startApp(t *testing.T, wrap func(http.Handler) http.Handler) (*board.Board, *httptest.Server) {
	t.Helper()
	b := board.New()
	h, err := httpapi.NewApp(b, web.Page())
	if err != nil {
		t.Fatal(err)
	}
	if wrap != nil {
		h = wrap(h)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return b, srv
}

func add(t *testing.T, b *board.Board, title string) board.Card {
	t.Helper()
	c, err := b.Add(title, "")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func moveTo(t *testing.T, b *board.Board, id string, to board.Column) {
	t.Helper()
	if _, err := b.Move(id, &to); err != nil {
		t.Fatalf("move %s to %s: %v", id, to, err)
	}
}

// serverColumn is the column the server holds card id in, or "" when absent.
func serverColumn(b *board.Board, id string) board.Column {
	for _, l := range b.Snapshot().Lanes {
		for _, c := range l.Cards {
			if c.ID == id {
				return l.Name
			}
		}
	}
	return ""
}

// --- the browser -------------------------------------------------------------

type browser struct {
	t   *testing.T
	ctx context.Context

	mu           sync.Mutex
	problems     []string
	allowStatus  []int // 4xx/5xx statuses of /api/* this test drives on purpose
	allowNetFail bool  // this test drives a network failure on purpose
}

// newBrowser starts a headless Chrome with a throwaway profile (chromedp's
// temporary user-data dir) for one test and checks its console at the end.
func newBrowser(t *testing.T) *browser {
	t.Helper()
	path, err := findChrome()
	if err != nil {
		t.Fatal(err)
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.ExecPath(path))
	actx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	bctx, cancelBrowser := chromedp.NewContext(actx)
	t.Cleanup(func() { cancelBrowser(); cancelAlloc() })

	// The first call starts the browser and must not carry a timeout.
	if err := chromedp.Do(bctx); err != nil {
		t.Fatalf("e2e: start Chrome at %s: %v", path, err)
	}
	ctx, cancelTimeout := context.WithTimeout(bctx, e2eTimeout)
	br := &browser{t: t, ctx: ctx}
	var wg sync.WaitGroup
	watch(&wg, bctx, runtime.ConsoleAPICalled, br.onConsole)
	watch(&wg, bctx, runtime.ExceptionThrown, br.onException)
	watch(&wg, bctx, cdplog.EntryAdded, br.onLog)
	watch(&wg, bctx, network.ResponseReceived, br.onResponse)
	watch(&wg, bctx, network.LoadingFailed, br.onLoadingFailed)
	// LIFO: check the console, then stop the watchers, before the browser goes.
	t.Cleanup(func() { cancelTimeout(); cancelBrowser(); wg.Wait() })
	t.Cleanup(br.assertClean)
	return br
}

// watch feeds every event ev of the tab to f until ctx ends.
func watch[E any](wg *sync.WaitGroup, ctx context.Context, ev cdp.Event[E], f func(E)) {
	events := chromedp.Events(ctx, ev)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for e, err := range events {
			if err != nil {
				return
			}
			f(e)
		}
	}()
}

func (br *browser) allow(statuses ...int) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.allowStatus = append(br.allowStatus, statuses...)
}

func (br *browser) allowNetworkFailure() {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.allowNetFail = true
}

func (br *browser) problem(format string, args ...any) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.problems = append(br.problems, fmt.Sprintf(format, args...))
}

func (br *browser) onConsole(e runtime.EventConsoleAPICalled) {
	switch e.Type {
	case runtime.ConsoleAPICalledTypeError, runtime.ConsoleAPICalledTypeWarning, runtime.ConsoleAPICalledTypeAssert:
		var parts []string
		for _, a := range e.Args {
			parts = append(parts, string(a.Value)+a.Description)
		}
		br.problem("console.%s: %s", e.Type, strings.Join(parts, " "))
	}
}

func (br *browser) onException(e runtime.EventExceptionThrown) {
	br.problem("uncaught %s", (&chromedp.ExceptionError{ExceptionDetails: e.ExceptionDetails}).Error())
}

// onLog: failed loads are judged by onResponse/onLoadingFailed with their
// status; every other warning or error (CSP violations, deprecations,
// interventions) is a problem.
func (br *browser) onLog(e cdplog.EventEntryAdded) {
	if e.Entry.Source != cdplog.EntrySourceNetwork &&
		(e.Entry.Level == cdplog.EntryLevelError || e.Entry.Level == cdplog.EntryLevelWarning) {
		br.problem("log %s/%s: %s", e.Entry.Source, e.Entry.Level, e.Entry.Text)
	}
}

func (br *browser) onResponse(e network.EventResponseReceived) {
	status := int(e.Response.Status)
	if status < 400 {
		return
	}
	br.mu.Lock()
	ok := isAPI(e.Response.URL) && slices.Contains(br.allowStatus, status)
	br.mu.Unlock()
	if !ok {
		br.problem("HTTP %d for %s", status, e.Response.URL)
	}
}

func (br *browser) onLoadingFailed(e network.EventLoadingFailed) {
	br.mu.Lock()
	ok := br.allowNetFail
	br.mu.Unlock()
	if !ok {
		br.problem("request failed: %s %s (canceled=%v)", e.Type, e.ErrorText, e.Canceled)
	}
}

func isAPI(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && strings.HasPrefix(u.Path, "/api/")
}

func (br *browser) assertClean() {
	br.mu.Lock()
	defer br.mu.Unlock()
	if len(br.problems) > 0 {
		br.t.Errorf("browser console is not clean:\n  %s", strings.Join(br.problems, "\n  "))
	}
}

func (br *browser) run(step string, actions ...chromedp.Action[chromedp.Void]) {
	br.t.Helper()
	if err := chromedp.Do(br.ctx, actions...); err != nil {
		br.t.Fatalf("%s: %v", step, err)
	}
}

// settled is true once no action is in flight and the board has been rendered.
const settled = `document.querySelector('#board').getAttribute('aria-busy') === 'false'`

// wait polls a JavaScript predicate until it holds (bounded by e2eTimeout).
func (br *browser) wait(step, predicate string) {
	br.t.Helper()
	br.run(step, chromedp.Poll[chromedp.Void](predicate, chromedp.WithPollingInterval(20*time.Millisecond), chromedp.WithPollingTimeout(0)))
}

func (br *browser) open(u string) {
	br.t.Helper()
	br.run("open "+u, chromedp.Navigate(u))
	br.wait("first render", settled)
}

func eval[T any](br *browser, step, js string) T {
	br.t.Helper()
	v, err := chromedp.Run(br.ctx, chromedp.Evaluate[T](js))
	if err != nil {
		br.t.Fatalf("%s: %v", step, err)
	}
	return v
}

// titles are the card titles the page shows in column name, top to bottom.
func (br *browser) titles(name string) []string {
	br.t.Helper()
	return eval[[]string](br, "read column "+name, fmt.Sprintf(
		`Array.from(document.querySelectorAll('[data-column=%q] .card-title')).map(e => e.textContent)`, name))
}

type notice struct {
	Hidden bool   `json:"hidden"`
	Code   string `json:"code"`
	Text   string `json:"text"`
	Role   string `json:"role"`
}

func (br *browser) notice() notice {
	br.t.Helper()
	return eval[notice](br, "read the refusal notice",
		`(() => { const n = document.querySelector('#refusal');
		  return {hidden: n.hidden, code: n.dataset.code || '', text: n.textContent, role: n.getAttribute('role')}; })()`)
}

func cardButton(id, action string) string {
	return fmt.Sprintf(`[data-card-id=%q] button[data-action=%q]`, id, action)
}

func (br *browser) click(sel string) {
	br.t.Helper()
	br.run("click "+sel, chromedp.Click(chromedp.CSS(sel)))
}

// expectRefusal waits for the re-render after a refused action and checks the
// RefusalNotice shows the reason for code.
func (br *browser) expectRefusal(code, reason string) notice {
	br.t.Helper()
	br.wait("refusal "+code, fmt.Sprintf(`document.querySelector('#refusal').dataset.code === %q && %s`, code, settled))
	n := br.notice()
	if n.Hidden || n.Role != "alert" || !strings.Contains(n.Text, reason) {
		br.t.Fatalf("notice = %+v, want visible role=alert containing %q", n, reason)
	}
	return n
}

func wantTitles(t *testing.T, br *browser, column string, want ...string) {
	t.Helper()
	if got := br.titles(column); !slices.Equal(got, want) && (len(got) != 0 || len(want) != 0) {
		t.Fatalf("column %s shows %q, want %q", column, got, want)
	}
}

// --- FR-010 ------------------------------------------------------------------

// AC-024
func TestE2E_PageShowsThreeColumnsWithCards(t *testing.T) {
	b, srv := startApp(t, nil)
	a := add(t, b, "A")
	bb := add(t, b, "B")
	c := add(t, b, "C")
	moveTo(t, b, bb.ID, board.Doing)
	moveTo(t, b, c.ID, board.Doing)
	moveTo(t, b, c.ID, board.Done)
	_ = a

	br := newBrowser(t)
	br.open(srv.URL)
	headings := eval[[]string](br, "read headings", `Array.from(document.querySelectorAll('#board section h2')).map(e => e.textContent)`)
	if !slices.Equal(headings, []string{"todo", "doing", "done"}) {
		t.Fatalf("column headings = %q", headings)
	}
	wantTitles(t, br, "todo", "A")
	wantTitles(t, br, "doing", "B")
	wantTitles(t, br, "done", "C")
	counts := eval[[]string](br, "read counts", `Array.from(document.querySelectorAll('#board .count')).map(e => e.textContent)`)
	if !slices.Equal(counts, []string{"1", "1", "1"}) {
		t.Fatalf("counts = %q", counts)
	}
	if n := br.notice(); !n.Hidden {
		t.Fatalf("notice shown on first load: %+v", n)
	}
}

// AC-025, plus the edge_controls_disabled state (ADR-0008).
func TestE2E_CardShowsFourControls(t *testing.T) {
	b, srv := startApp(t, nil)
	a := add(t, b, "A")
	d := add(t, b, "D")
	moveTo(t, b, d.ID, board.Doing)
	moveTo(t, b, d.ID, board.Done)

	br := newBrowser(t)
	br.open(srv.URL)
	type control struct {
		Label    string `json:"label"`
		Disabled bool   `json:"disabled"`
		Icon     bool   `json:"icon"`
		Text     string `json:"text"`
	}
	controls := func(id string) []control {
		return eval[[]control](br, "read controls of "+id, fmt.Sprintf(
			`Array.from(document.querySelectorAll('[data-card-id=%q] .card-actions button')).map(b => (
			  {label: b.getAttribute('aria-label'), disabled: b.disabled, icon: !!b.querySelector('svg use'), text: b.textContent.trim()}))`, id))
	}
	want := func(id string, disabled ...bool) {
		t.Helper()
		got := controls(id)
		labels := []string{"Move left", "Move right", "Edit", "Delete"}
		if len(got) != len(labels) {
			t.Fatalf("card %s has %d controls: %+v", id, len(got), got)
		}
		for i, c := range got {
			if c.Label != labels[i] || c.Disabled != disabled[i] || !c.Icon || c.Text != "" {
				t.Fatalf("card %s control %d = %+v, want %q disabled=%v with an SVG icon and no text", id, i, c, labels[i], disabled[i])
			}
		}
	}
	want(a.ID, true, false, false, false) // todo: move left disabled
	want(d.ID, false, true, false, false) // done: move right disabled
}

// AC-026, plus the empty and add_form states.
func TestE2E_AddFormCreatesCardInTodo(t *testing.T) {
	b, srv := startApp(t, nil)
	br := newBrowser(t)
	br.open(srv.URL)

	empty := eval[[]string](br, "read empty columns", `Array.from(document.querySelectorAll('#board .column-empty')).map(e => e.textContent)`)
	if !slices.Equal(empty, []string{"No cards", "No cards", "No cards"}) {
		t.Fatalf("empty board shows %q", empty)
	}
	br.run("type a title", chromedp.SendKeys(chromedp.CSS("#add-title"), "Ship it"))
	br.click("#add-submit")
	br.wait("card rendered", `Array.from(document.querySelectorAll('[data-column="todo"] .card-title')).some(e => e.textContent === 'Ship it') && `+settled)

	wantTitles(t, br, "todo", "Ship it")
	if n := br.notice(); !n.Hidden {
		t.Fatalf("notice after a successful add: %+v", n)
	}
	if v := eval[string](br, "read title input", `document.querySelector('#add-title').value`); v != "" {
		t.Fatalf("title input keeps %q after a successful add", v)
	}
	lanes := b.Snapshot().Lanes
	if len(lanes[0].Cards) != 1 || lanes[0].Cards[0].Title != "Ship it" {
		t.Fatalf("server todo = %+v", lanes[0].Cards)
	}
}

// --- FR-011: refusal reasons -------------------------------------------------

// AC-027, plus the doing_at_wip_limit state.
func TestE2E_WipRefusalShownOnPage(t *testing.T) {
	b, srv := startApp(t, nil)
	for _, title := range []string{"D1", "D2", "D3"} {
		moveTo(t, b, add(t, b, title).ID, board.Doing)
	}
	a := add(t, b, "A")

	br := newBrowser(t)
	br.allow(http.StatusConflict)
	br.open(srv.URL)
	if !eval[bool](br, "read WIP outline", `document.querySelector('[data-column="doing"]').classList.contains('column-at-limit')`) {
		t.Fatal("doing at 3 cards has no WIP-limit outline")
	}
	br.click(cardButton(a.ID, "right"))
	br.expectRefusal("wip_limit_exceeded", "doing already holds 3 cards")

	wantTitles(t, br, "todo", "A")
	wantTitles(t, br, "doing", "D1", "D2", "D3")
	if col := serverColumn(b, a.ID); col != board.Todo {
		t.Fatalf("server has A in %q", col)
	}
}

// AC-028
func TestE2E_InvalidTitleReasonShownOnPage(t *testing.T) {
	b, srv := startApp(t, nil)
	br := newBrowser(t)
	br.allow(http.StatusBadRequest)
	br.open(srv.URL)

	br.click("#add-submit") // blank title
	br.expectRefusal("validation_failed", "a title is required")

	wantTitles(t, br, "todo")
	if lanes := b.Snapshot().Lanes; len(lanes[0].Cards) != 0 {
		t.Fatalf("a card was added: %+v", lanes[0].Cards)
	}
	if v := eval[string](br, "read aria-invalid", `document.querySelector('#add-title').getAttribute('aria-invalid')`); v != "true" {
		t.Fatalf("title input aria-invalid = %q", v)
	}
}

// AC-034: the page is stale, the click sends to 'done'.
func TestE2E_InvalidMoveReasonShownOnPage(t *testing.T) {
	b, srv := startApp(t, nil)
	a := add(t, b, "A")
	moveTo(t, b, a.ID, board.Doing)

	br := newBrowser(t)
	br.allow(http.StatusConflict)
	br.open(srv.URL)
	wantTitles(t, br, "doing", "A")

	moveTo(t, b, a.ID, board.Todo) // another member, no polling (ASM-0004)
	br.click(cardButton(a.ID, "right"))
	br.expectRefusal("invalid_move", "no longer where this page showed it")

	wantTitles(t, br, "todo", "A")
	wantTitles(t, br, "doing")
	if col := serverColumn(b, a.ID); col != board.Todo {
		t.Fatalf("server has A in %q", col)
	}
}

// AC-035
func TestE2E_NotFoundReasonShownOnStalePage(t *testing.T) {
	b, srv := startApp(t, nil)
	a := add(t, b, "A")

	br := newBrowser(t)
	br.allow(http.StatusNotFound)
	br.open(srv.URL)
	wantTitles(t, br, "todo", "A")

	if err := b.Delete(a.ID); err != nil { // another member
		t.Fatal(err)
	}
	br.click(cardButton(a.ID, "right"))
	br.expectRefusal("not_found", "no longer exists")
	wantTitles(t, br, "todo")
}

// --- the remaining FR-010 states and the declared failure modes --------------

// edit_mode, edit_saved, card_moved, card_deleted, a refused save, and the
// one-action-in-flight rule (two synchronous clicks send one move).
func TestE2E_EditSaveMoveAndDeleteFlow(t *testing.T) {
	b, srv := startApp(t, nil)
	a := add(t, b, "Write docs")
	other := add(t, b, "Other")
	last := add(t, b, "Last")

	br := newBrowser(t)
	br.allow(http.StatusBadRequest)
	br.open(srv.URL)

	// edit_mode: the form replaces the card, pre-filled.
	br.click(cardButton(a.ID, "edit"))
	br.wait("edit form", `document.querySelector('#edit-card') !== null`)
	if v := eval[string](br, "read edit title", `document.querySelector('#edit-title').value`); v != "Write docs" {
		t.Fatalf("edit form title = %q", v)
	}
	// A refused save keeps the form open with the draft and marks the title.
	br.run("clear the title", chromedp.SetValue(chromedp.CSS("#edit-title"), "   "))
	br.click("#edit-save")
	br.expectRefusal("validation_failed", "a title is required")
	if v := eval[string](br, "read edit state", `(() => { const i = document.querySelector('#edit-title'); return i ? i.value + '|' + i.getAttribute('aria-invalid') : 'closed'; })()`); v != "   |true" {
		t.Fatalf("after a refused save the edit form is %q", v)
	}
	// edit_saved
	br.run("type the new title", chromedp.SetValue(chromedp.CSS("#edit-title"), "Write the docs"))
	br.click("#edit-save")
	br.wait("saved", `document.querySelector('#edit-card') === null && document.querySelector('#refusal').hidden && `+settled)
	wantTitles(t, br, "todo", "Write the docs", "Other", "Last")

	// card_moved, with a second click while the first is in flight.
	br.run("double move", chromedp.Evaluate[chromedp.Void](fmt.Sprintf(
		`(() => { const b = document.querySelector(%q); b.click(); b.click(); return true; })()`, cardButton(a.ID, "right"))))
	br.wait("moved", `document.querySelectorAll('[data-column="doing"] .card').length === 1 && `+settled)
	wantTitles(t, br, "doing", "Write the docs")
	if col := serverColumn(b, a.ID); col != board.Doing {
		t.Fatalf("two quick clicks left the card in %q, want doing (one move)", col)
	}
	if n := br.notice(); !n.Hidden {
		t.Fatalf("notice after a move: %+v", n)
	}

	// card_deleted: the rest stays dense (INV-004).
	br.click(cardButton(other.ID, "delete"))
	br.wait("deleted", `!document.querySelector('[data-card-id="`+other.ID+`"]') && `+settled)
	wantTitles(t, br, "todo", "Last")
	if lanes := b.Snapshot().Lanes; len(lanes[0].Cards) != 1 || lanes[0].Cards[0].ID != last.ID || lanes[0].Cards[0].Position != 0 {
		t.Fatalf("server todo after delete = %+v", lanes[0].Cards)
	}
}

// Failure matrix: the server is gone; the page says so and keeps the board.
func TestE2E_NetworkErrorReasonShown(t *testing.T) {
	b, srv := startApp(t, nil)
	a := add(t, b, "A")
	br := newBrowser(t)
	br.allowNetworkFailure()
	br.open(srv.URL)

	srv.Close()
	br.click(cardButton(a.ID, "right"))
	br.expectRefusal("network", "the board could not be reached")
	wantTitles(t, br, "todo", "A")
}

// Failure matrix: a response slower than REQUEST_TIMEOUT_MS is aborted and
// reported like a network failure; the re-read then shows the real board.
func TestE2E_SlowResponseTimesOutWithReason(t *testing.T) {
	b, srv := startApp(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/move") {
				select { // answer only after the page has given up
				case <-r.Context().Done():
				case <-time.After(requestTimeout + time.Second):
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	a := add(t, b, "A")
	br := newBrowser(t)
	br.allowNetworkFailure()
	br.open(srv.URL)

	start := time.Now()
	br.click(cardButton(a.ID, "right"))
	n := br.expectRefusal("network", "Cannot move “A”: the board could not be reached")
	if took := time.Since(start); took < requestTimeout || took > 2*requestTimeout {
		t.Fatalf("timed out after %v, want within [%v, %v] (%+v)", took, requestTimeout, 2*requestTimeout, n)
	}
	wantTitles(t, br, "todo", "A")
}

// The reason mapping and render branches the real server cannot produce:
// unknown codes, non-JSON and malformed error bodies, a non-JSON board.
func TestE2E_PageLogicBranches(t *testing.T) {
	var mu sync.Mutex
	var stub struct {
		status int
		ctype  string
		body   string
		board  string // non-empty: the body of GET /api/board
	}
	set := func(status int, ctype, body, boardBody string) {
		mu.Lock()
		defer mu.Unlock()
		stub.status, stub.ctype, stub.body, stub.board = status, ctype, body, boardBody
	}
	b, srv := startApp(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			s := stub
			mu.Unlock()
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/api/board" && s.board != "":
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, s.board)
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/move") && s.status != 0:
				w.Header().Set("Content-Type", s.ctype)
				w.WriteHeader(s.status)
				_, _ = io.WriteString(w, s.body)
			default:
				next.ServeHTTP(w, r)
			}
		})
	})
	a := add(t, b, "A")
	for _, title := range []string{"D1", "D2"} {
		moveTo(t, b, add(t, b, title).ID, board.Doing)
	}
	br := newBrowser(t)
	br.allow(400, 404, 409, 500, 502)
	br.open(srv.URL)
	if eval[bool](br, "read WIP outline", `document.querySelector('[data-column="doing"]').classList.contains('column-at-limit')`) {
		t.Fatal("doing at 2 cards has the WIP-limit outline")
	}

	cases := []struct {
		status      int
		ctype, body string
		code, text  string
	}{
		{500, "text/plain; charset=utf-8", "internal error\n", "unexpected", "unexpected response (HTTP 500)"},
		{502, "application/json", `{"code":42}`, "unexpected", "unexpected response (HTTP 502)"},
		{409, "application/json", `{"code":"new_rule"}`, "new_rule", `refused it (code "new_rule")`},
		{409, "application/json", `{"code":"wip_limit_exceeded"}`, "wip_limit_exceeded", "doing already holds 3 cards"},
		{409, "application/json", `{"code":"invalid_move"}`, "invalid_move", "no longer where this page showed it"},
		{404, "application/json", `{"code":"not_found"}`, "not_found", "no longer exists"},
		{400, "application/json", `{"code":"validation_failed"}`, "validation_failed", "a title is required"},
	}
	for _, c := range cases {
		set(c.status, c.ctype, c.body, "")
		br.run("clear notice", chromedp.Evaluate[chromedp.Void](`(() => { const n = document.querySelector('#refusal'); n.hidden = true; delete n.dataset.code; return true; })()`))
		br.click(cardButton(a.ID, "right"))
		br.expectRefusal(c.code, "Cannot move “A”: ")
		if n := br.notice(); !strings.Contains(n.Text, c.text) {
			t.Errorf("%d %s: notice %q, want %q", c.status, c.body, n.Text, c.text)
		}
		wantTitles(t, br, "todo", "A") // re-read after every refusal
	}

	// A board that is not JSON on first load: the reason, no exception.
	set(0, "", "", "not json")
	br.open(srv.URL)
	if n := br.notice(); n.Code != "unexpected" || !strings.Contains(n.Text, "Cannot load the board") {
		t.Fatalf("non-JSON board: notice %+v", n)
	}
}

// --- NFR-003: the built binary -----------------------------------------------

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "board")
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

// startBinary runs the binary with PORT set to port, or unset when port is "",
// and returns the base URL once GET / answers. It refuses a port that is
// already taken, so it never tests somebody else's server.
func startBinary(t *testing.T, bin, port, listenPort string) string {
	t.Helper()
	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		t.Fatalf("port %s is in use (%v); free it to run this test", listenPort, err)
	}
	_ = ln.Close()

	env := slices.DeleteFunc(os.Environ(), func(kv string) bool { return strings.HasPrefix(kv, "PORT=") })
	if port != "" {
		env = append(env, "PORT="+port)
	}
	cmd := exec.Command(bin)
	cmd.Env = env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case err := <-exited:
			if err != nil {
				t.Errorf("binary exit after SIGTERM: %v\n%s", err, stderr.String())
			}
		case <-time.After(shutdownTimeout + time.Second):
			_ = cmd.Process.Kill()
			t.Errorf("binary did not stop within %v of SIGTERM", shutdownTimeout+time.Second)
		}
	})

	base := "http://127.0.0.1:" + listenPort
	deadline := time.Now().Add(startTimeout)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case err := <-exited:
			t.Fatalf("binary exited before serving: %v\n%s", err, stderr.String())
		case <-tick.C:
		}
		if resp, err := binaryClient.Get(base + "/"); err == nil {
			_ = resp.Body.Close()
			return base
		}
		if time.Now().After(deadline) {
			t.Fatalf("binary did not answer on port %s within %v\n%s", listenPort, startTimeout, stderr.String())
		}
	}
}

// binaryClient opens one connection per request and closes it, so the test
// leaves no idle or never-used connection behind for the binary's graceful
// shutdown to wait on.
var binaryClient = &http.Client{Timeout: time.Second, Transport: &http.Transport{DisableKeepAlives: true}}

func wantBoardPage(t *testing.T, base string) {
	t.Helper()
	for path, marker := range map[string]string{
		"/":                  `<title>Team board</title>`,
		"/static/app.js":     "REQUEST_TIMEOUT_MS",
		"/static/tokens.css": "--color-accent",
	} {
		resp, err := binaryClient.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), marker) {
			t.Fatalf("GET %s = %d, body without %q", path, resp.StatusCode, marker)
		}
	}
}

// AC-031
func TestE2E_BinaryServesPageOnDefaultPort(t *testing.T) {
	base := startBinary(t, buildBinary(t), "", defaultPort)
	wantBoardPage(t, base)
}

// AC-032, and the wave-3 checkpoint: the page on PORT=9090 renders in a
// browser with a clean console.
func TestE2E_BinaryHonoursPortVariable(t *testing.T) {
	base := startBinary(t, buildBinary(t), "9090", "9090")
	wantBoardPage(t, base)

	br := newBrowser(t)
	br.open(base)
	headings := eval[[]string](br, "read headings", `Array.from(document.querySelectorAll('#board section h2')).map(e => e.textContent)`)
	if !slices.Equal(headings, []string{"todo", "doing", "done"}) {
		t.Fatalf("column headings = %q", headings)
	}
}
