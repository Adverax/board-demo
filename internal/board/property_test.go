package board_test

// Property tests, one per invariant of AGG-001 this card touches:
//   INV-001 title 1..80 runes    -> TestPropertyTest_Title_ValidOnAddAndEdit, and TestPropertyTest_DenseOrder_AnyOperationSequence (model rejects blank and over-long titles on add and edit)
//   INV-002 doing <= WIPLimit    -> TestPropertyTest_Doing_NeverExceedsWIPUnderConcurrency, and checked after every step of the model test
//   INV-003 adjacent moves only  -> TestPropertyTest_DenseOrder_AnyOperationSequence (model refuses non-neighbouring targets)
//   INV-004 dense order          -> TestPropertyTest_DenseOrder_RandomOperations, TestPropertyTest_DenseOrder_AnyOperationSequence
//   INV-005 three fixed columns  -> denseViolations, checked after every step
//   ADR-0006/R1 lock + copy      -> TestPropertyTest_Board_ConcurrentOpsSnapshotIsCopy
//
// Go requires test functions to start with "Test"; the traced names
// (PropertyTest_...) follow the prefix verbatim.

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"pgregory.net/rapid"

	"github.com/adverax/board-demo/internal/board"
)

var allColumns = []board.Column{board.Todo, board.Doing, board.Done}

// targets includes every outcome class of a move target: the three columns,
// an unknown name, the empty name (a named target, not an absent one) and no
// readable target (nil).
var targets = []*board.Column{target(board.Todo), target(board.Doing), target(board.Done), target("archive"), target(""), nil}

var titles = []string{"a", "  padded  ", "", "   ", "Write docs",
	strings.Repeat("x", board.MaxTitleLength), strings.Repeat("я", board.MaxTitleLength+1)}

// wantTitle is the reference INV-001 rule: the trimmed title, valid when it
// is 1..MaxTitleLength runes long (ASM-0002).
func wantTitle(s string) (string, bool) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return s, n > 0 && n <= board.MaxTitleLength
}

// EC(INV-001) / FR-008: for any title, Add and Edit accept it exactly when the
// trimmed title is 1..80 runes, store the trimmed value, and a refusal changes
// nothing (Edit keeps title, description, column and position).
func TestPropertyTest_Title_ValidOnAddAndEdit(t *testing.T) {
	ws := rapid.SampledFrom([]string{"", " ", "\t", "\n", "  \t "})
	body := rapid.OneOf(
		rapid.String(),
		rapid.StringOfN(rapid.RuneFrom([]rune("aя€😀")), board.MaxTitleLength-2, board.MaxTitleLength+2, -1),
	)
	rapid.Check(t, func(t *rapid.T) {
		title := ws.Draw(t, "lead") + body.Draw(t, "body") + ws.Draw(t, "trail")
		want, ok := wantTitle(title)

		b := board.New()
		c, err := b.Add(title, "")
		if ok != (err == nil) || (!ok && !errors.Is(err, board.ErrValidation)) {
			t.Fatalf("Add(%q) err = %v, want ok=%v", title, err, ok)
		}
		if ok && c.Title != want {
			t.Fatalf("Add stored %q, want %q", c.Title, want)
		}
		wantCards := 0
		if ok {
			wantCards = 1
		}
		if n := len(b.Snapshot().Lanes[0].Cards); n != wantCards {
			t.Fatalf("todo holds %d cards after Add, want %d", n, wantCards)
		}

		ids := fill(t, b, 2)
		before := b.Snapshot()
		got, err := b.Edit(ids[0], &title, nil)
		if ok != (err == nil) || (!ok && !errors.Is(err, board.ErrValidation)) {
			t.Fatalf("Edit(%q) err = %v, want ok=%v", title, err, ok)
		}
		if !ok {
			if !snapshotsEqual(before, b.Snapshot()) {
				t.Fatalf("refused edit changed the board")
			}
			return
		}
		old := before.Lanes[0].Cards[len(before.Lanes[0].Cards)-2]
		if got.Title != want || got.Description != old.Description || got.Column != old.Column || got.Position != old.Position {
			t.Fatalf("Edit returned %+v, want title %q in place of %+v", got, want, old)
		}
	})
}

func allIDs(s board.Snapshot) []string {
	var ids []string
	for _, l := range s.Lanes {
		for _, c := range l.Cards {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

// pickID draws an existing card id or, sometimes, an unknown one.
func pickID(t *rapid.T, b *board.Board) string {
	ids := append(allIDs(b.Snapshot()), "missing")
	return rapid.SampledFrom(ids).Draw(t, "id")
}

func optional(t *rapid.T, label string, values []string) *string {
	if !rapid.Bool().Draw(t, label+"_set") {
		return nil
	}
	v := rapid.SampledFrom(values).Draw(t, label)
	return &v
}

// randomActions are add, move, edit and delete with arbitrary (valid and
// invalid) arguments; their results are deliberately ignored.
func randomActions(b *board.Board) map[string]func(*rapid.T) {
	return map[string]func(*rapid.T){
		"add": func(t *rapid.T) {
			_, _ = b.Add(rapid.SampledFrom(titles).Draw(t, "title"), "d")
		},
		"move": func(t *rapid.T) {
			_, _ = b.Move(pickID(t, b), rapid.SampledFrom(targets).Draw(t, "to"))
		},
		"edit": func(t *rapid.T) {
			_, _ = b.Edit(pickID(t, b), optional(t, "title", titles), optional(t, "desc", titles))
		},
		"delete": func(t *rapid.T) {
			_ = b.Delete(pickID(t, b))
		},
	}
}

// AC-030 / NFR-002: dense_order_violations == 0 after every operation of a
// random add/move/edit/delete sequence applied to an empty board.
func TestPropertyTest_DenseOrder_RandomOperations(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		b := board.New()
		actions := randomActions(b)
		actions[""] = func(t *rapid.T) {
			if v := denseViolations(b.Snapshot()); v != 0 {
				t.Fatalf("dense_order_violations = %d", v)
			}
		}
		t.Repeat(actions)
	})
}

// model is a reference implementation of the board rules: card ids per lane.
type model struct {
	lanes [3][]string
	title map[string]string
}

func (m *model) locate(id string) (ci, pos int, ok bool) {
	for ci, lane := range m.lanes {
		if pos := slices.Index(lane, id); pos >= 0 {
			return ci, pos, true
		}
	}
	return 0, 0, false
}

// EC-002 / INV-004: after any sequence of add, move, edit and delete, the
// board agrees with a reference model and every column's positions are 0..n-1.
// Refusals are checked too: a refused operation returns the model's error and
// changes nothing.
func TestPropertyTest_DenseOrder_AnyOperationSequence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		b := board.New()
		m := &model{title: map[string]string{}}
		trimmed := wantTitle
		expectErr := func(t *rapid.T, op string, got, want error) {
			if !errors.Is(got, want) || (want == nil && got != nil) {
				t.Fatalf("%s: err = %v, want %v", op, got, want)
			}
		}

		t.Repeat(map[string]func(*rapid.T){
			"add": func(t *rapid.T) {
				title := rapid.SampledFrom(titles).Draw(t, "title")
				c, err := b.Add(title, "")
				tt, ok := trimmed(title)
				if !ok {
					expectErr(t, "add", err, board.ErrValidation)
					return
				}
				expectErr(t, "add", err, nil)
				if c.Title != tt || c.Column != board.Todo || c.Position != len(m.lanes[0]) {
					t.Fatalf("add returned %+v", c)
				}
				m.lanes[0] = append(m.lanes[0], c.ID)
				m.title[c.ID] = tt
			},
			"move": func(t *rapid.T) {
				id := pickID(t, b)
				to := rapid.SampledFrom(targets).Draw(t, "to")
				c, err := b.Move(id, to)
				from, pos, ok := m.locate(id)
				switch {
				case !ok:
					expectErr(t, "move", err, board.ErrNotFound)
				case to == nil:
					expectErr(t, "move", err, board.ErrValidation)
				case slices.Index(allColumns, *to) < 0 || (slices.Index(allColumns, *to) != from-1 && slices.Index(allColumns, *to) != from+1):
					expectErr(t, "move", err, board.ErrInvalidMove)
				case *to == board.Doing && len(m.lanes[1]) >= board.WIPLimit:
					expectErr(t, "move", err, board.ErrWIPLimit)
				default:
					expectErr(t, "move", err, nil)
					ti := slices.Index(allColumns, *to)
					m.lanes[from] = slices.Delete(m.lanes[from], pos, pos+1)
					m.lanes[ti] = append(m.lanes[ti], id)
					if c.Column != *to || c.Position != len(m.lanes[ti])-1 {
						t.Fatalf("move returned %+v", c)
					}
				}
			},
			"edit": func(t *rapid.T) {
				id := pickID(t, b)
				title := optional(t, "title", titles)
				_, err := b.Edit(id, title, nil)
				_, _, ok := m.locate(id)
				switch {
				case !ok:
					expectErr(t, "edit", err, board.ErrNotFound)
				case title == nil:
					expectErr(t, "edit", err, nil)
				default:
					tt, valid := trimmed(*title)
					if !valid {
						expectErr(t, "edit", err, board.ErrValidation)
						return
					}
					expectErr(t, "edit", err, nil)
					m.title[id] = tt
				}
			},
			"delete": func(t *rapid.T) {
				id := pickID(t, b)
				err := b.Delete(id)
				ci, pos, ok := m.locate(id)
				if !ok {
					expectErr(t, "delete", err, board.ErrNotFound)
					return
				}
				expectErr(t, "delete", err, nil)
				m.lanes[ci] = slices.Delete(m.lanes[ci], pos, pos+1)
				delete(m.title, id)
			},
			"": func(t *rapid.T) {
				s := b.Snapshot()
				if v := denseViolations(s); v != 0 {
					t.Fatalf("dense_order_violations = %d: %+v", v, s)
				}
				if n := len(s.Lanes[1].Cards); n > board.WIPLimit {
					t.Fatalf("doing holds %d cards", n)
				}
				for ci, l := range s.Lanes {
					if got := laneIDs(s, l.Name); !equal(got, m.lanes[ci]) {
						t.Fatalf("lane %s = %v, model %v", l.Name, got, m.lanes[ci])
					}
					for _, c := range l.Cards {
						if c.Title != m.title[c.ID] {
							t.Fatalf("card %s title %q, model %q", c.ID, c.Title, m.title[c.ID])
						}
					}
				}
			},
		})
	})
}

// runConcurrently starts every op at once (behind a barrier) and, while they
// run, polls snapshots. It returns every polled snapshot plus one taken after
// all ops settled.
func runConcurrently(b *board.Board, ops []func()) (polled []board.Snapshot, settled board.Snapshot) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, op := range ops {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			op()
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	close(start)
	for {
		select {
		case <-done:
			return polled, b.Snapshot()
		default:
			polled = append(polled, b.Snapshot())
		}
	}
}

// AC-029 / NFR-001: max_doing_cards_under_concurrent_moves <= 3 for 20
// simultaneous moves into an empty doing, observed by polling snapshots while
// the moves run and on the settled board (ADR-0005).
func TestConcurrency_20SimultaneousMovesIntoDoing(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 20)
	ops := make([]func(), len(ids))
	for i, id := range ids {
		ops[i] = func() { _, _ = b.Move(id, target(board.Doing)) }
	}
	polled, settled := runConcurrently(b, ops)
	maxDoing := 0
	for _, s := range append(polled, settled) {
		maxDoing = max(maxDoing, len(s.Lanes[1].Cards))
	}
	if maxDoing > board.WIPLimit {
		t.Fatalf("max_doing_cards_under_concurrent_moves = %d, want <= %d", maxDoing, board.WIPLimit)
	}
	if n := len(settled.Lanes[1].Cards); n != board.WIPLimit {
		t.Fatalf("settled doing holds %d, want %d", n, board.WIPLimit)
	}
}

// EC-001 / INV-002: for any number of simultaneous moves into doing (from todo
// and from done, mixed with moves out of doing), every snapshot read and the
// settled board show at most WIPLimit cards in doing, and no card is lost or
// duplicated.
func TestPropertyTest_Doing_NeverExceedsWIPUnderConcurrency(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		b := board.New()
		inDoing := rapid.IntRange(0, board.WIPLimit).Draw(t, "in_doing")
		inDone := rapid.IntRange(0, 5).Draw(t, "in_done")
		inTodo := rapid.IntRange(1, 30).Draw(t, "in_todo")
		// Build the starting board through the public API.
		for range inDone {
			c, _ := b.Add("x", "")
			_, _ = b.Move(c.ID, target(board.Doing))
			_, _ = b.Move(c.ID, target(board.Done))
		}
		for range inDoing {
			c, _ := b.Add("x", "")
			_, _ = b.Move(c.ID, target(board.Doing))
		}
		for range inTodo {
			_, _ = b.Add("x", "")
		}
		start := b.Snapshot()
		total := len(allIDs(start))

		var ops []func()
		for _, l := range start.Lanes {
			for _, c := range l.Cards {
				to := board.Doing
				if l.Name == board.Doing {
					to = rapid.SampledFrom([]board.Column{board.Todo, board.Done}).Draw(t, "out_of_doing")
				}
				if rapid.Bool().Draw(t, "include") {
					ops = append(ops, func() { _, _ = b.Move(c.ID, &to) })
				}
			}
		}
		polled, settled := runConcurrently(b, ops)
		for _, s := range append(polled, settled) {
			if n := len(s.Lanes[1].Cards); n > board.WIPLimit {
				t.Fatalf("a snapshot shows %d cards in doing", n)
			}
			if v := denseViolations(s); v != 0 {
				t.Fatalf("dense_order_violations = %d", v)
			}
		}
		ids := allIDs(settled)
		slices.Sort(ids)
		if len(ids) != total || len(slices.Compact(ids)) != total {
			t.Fatalf("cards lost or duplicated: %d -> %v", total, ids)
		}
	})
}

// ADR-0006/R1: under concurrent Snapshot/Add/Move/Edit/Delete, (a) a snapshot
// taken earlier never changes, (b) mutating a snapshot never reaches the
// board, (c) no update is lost: the settled board holds exactly the cards
// whose operations succeeded, and (d) concurrent edits, valid or refused,
// leave only INV-001-valid titles.
func TestPropertyTest_Board_ConcurrentOpsSnapshotIsCopy(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		b := board.New()
		seed := fill(t, b, rapid.IntRange(1, 8).Draw(t, "seed"))
		before := b.Snapshot()
		frozen := deepCopy(before)

		var mu sync.Mutex
		added, deleted := 0, 0
		n := rapid.IntRange(1, 40).Draw(t, "ops")
		ops := make([]func(), n)
		for i := range ops {
			id := rapid.SampledFrom(seed).Draw(t, "id")
			switch rapid.IntRange(0, 4).Draw(t, "kind") {
			case 0:
				ops[i] = func() {
					if _, err := b.Add("new", ""); err == nil {
						mu.Lock()
						added++
						mu.Unlock()
					}
				}
			case 1:
				to := rapid.SampledFrom(allColumns).Draw(t, "to")
				ops[i] = func() { _, _ = b.Move(id, &to) }
			case 2:
				title := optional(t, "title", titles)
				desc := optional(t, "desc", titles)
				ops[i] = func() { _, _ = b.Edit(id, title, desc) }
			case 3:
				ops[i] = func() {
					if b.Delete(id) == nil {
						mu.Lock()
						deleted++
						mu.Unlock()
					}
				}
			default:
				ops[i] = func() {
					s := b.Snapshot() // caller mutation of its own copy
					for li := range s.Lanes {
						for ci := range s.Lanes[li].Cards {
							s.Lanes[li].Cards[ci].Title = "mutated"
						}
						s.Lanes[li].Cards = nil
					}
				}
			}
		}
		_, settled := runConcurrently(b, ops)

		if !snapshotsEqual(before, frozen) {
			t.Fatalf("an earlier snapshot changed under later operations")
		}
		got := 0
		for _, l := range settled.Lanes {
			for _, c := range l.Cards {
				if c.Title == "mutated" {
					t.Fatalf("snapshot mutation reached the board: %+v", c)
				}
				if tt, ok := wantTitle(c.Title); !ok || tt != c.Title {
					t.Fatalf("settled card has an invalid title: %+v", c)
				}
				got++
			}
		}
		if want := len(seed) + added - deleted; got != want {
			t.Fatalf("settled board holds %d cards, want %d (lost update)", got, want)
		}
		if v := denseViolations(settled); v != 0 || len(settled.Lanes[1].Cards) > board.WIPLimit {
			t.Fatalf("settled board broke an invariant: %+v", settled)
		}
	})
}

func deepCopy(s board.Snapshot) board.Snapshot {
	for i := range s.Lanes {
		s.Lanes[i].Cards = slices.Clone(s.Lanes[i].Cards)
	}
	return s
}

func snapshotsEqual(a, b board.Snapshot) bool {
	for i := range a.Lanes {
		if a.Lanes[i].Name != b.Lanes[i].Name || !slices.Equal(a.Lanes[i].Cards, b.Lanes[i].Cards) {
			return false
		}
	}
	return true
}
