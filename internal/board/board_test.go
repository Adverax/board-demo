package board_test

import (
	"errors"
	"testing"

	"github.com/adverax/board-demo/internal/board"
)

// fataler is the part of testing.TB (and *rapid.T) that fill needs.
type fataler interface {
	Helper()
	Fatalf(format string, args ...any)
}

// fill adds n cards to todo and returns their ids in order.
func fill(t fataler, b *board.Board, n int) []string {
	t.Helper()
	ids := make([]string, n)
	for i := range ids {
		c, err := b.Add("card", "")
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		ids[i] = c.ID
	}
	return ids
}

// target returns a pointer to a named move target.
func target(c board.Column) *board.Column { return &c }

func mustMove(t *testing.T, b *board.Board, id string, to board.Column) {
	t.Helper()
	if _, err := b.Move(id, &to); err != nil {
		t.Fatalf("Move(%s, %s): %v", id, to, err)
	}
}

func laneIDs(s board.Snapshot, c board.Column) []string {
	for _, l := range s.Lanes {
		if l.Name == c {
			ids := make([]string, len(l.Cards))
			for i, card := range l.Cards {
				ids[i] = card.ID
			}
			return ids
		}
	}
	return nil
}

func assertDense(t *testing.T, s board.Snapshot) {
	t.Helper()
	if v := denseViolations(s); v != 0 {
		t.Fatalf("dense order violated (%d violations): %+v", v, s)
	}
}

// denseViolations counts cards whose Position or Column disagrees with their
// place in the snapshot (INV-004), plus a wrong column order (INV-005).
func denseViolations(s board.Snapshot) int {
	v := 0
	want := []board.Column{board.Todo, board.Doing, board.Done}
	for i, l := range s.Lanes {
		if l.Name != want[i] {
			v++
		}
		for pos, c := range l.Cards {
			if c.Position != pos || c.Column != l.Name {
				v++
			}
		}
	}
	return v
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestBoard_NewHasThreeEmptyColumnsInOrder(t *testing.T) {
	s := board.New().Snapshot()
	want := []board.Column{board.Todo, board.Doing, board.Done}
	for i, l := range s.Lanes {
		if l.Name != want[i] || len(l.Cards) != 0 {
			t.Fatalf("lane %d = %+v, want empty %s", i, l, want[i])
		}
	}
}

func TestAdd_TrimsTitleAndLandsAtEndOfTodo(t *testing.T) {
	b := board.New()
	fill(t, b, 2)
	c, err := b.Add("  Write docs  ", "draft")
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Write docs" || c.Description != "draft" || c.Column != board.Todo || c.Position != 2 {
		t.Fatalf("card = %+v", c)
	}
}

func TestAdd_RejectsBlankTitle(t *testing.T) {
	b := board.New()
	fill(t, b, 2)
	for _, title := range []string{"", "   ", "\t\n"} {
		if _, err := b.Add(title, "x"); !errors.Is(err, board.ErrValidation) {
			t.Fatalf("Add(%q) err = %v, want ErrValidation", title, err)
		}
	}
	if n := len(b.Snapshot().Lanes[0].Cards); n != 2 {
		t.Fatalf("todo holds %d cards, want 2", n)
	}
}

func TestAdd_IDsAreUnique(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 50)
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
}

func TestMove_Refusals(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 5) // todo = 1..5
	mustMove(t, b, ids[0], board.Doing)
	mustMove(t, b, ids[0], board.Done) // done=[1]
	mustMove(t, b, ids[1], board.Doing)
	mustMove(t, b, ids[2], board.Doing)
	mustMove(t, b, ids[3], board.Doing) // doing=[2 3 4] full, todo=[5]
	inTodo, inDoing, inDone := ids[4], ids[1], ids[0]

	cases := []struct {
		name string
		id   string
		to   *board.Column
		want error
	}{
		{"unknown id", "missing", target(board.Doing), board.ErrNotFound},
		{"unknown id beats no target", "missing", nil, board.ErrNotFound},
		{"no target", inTodo, nil, board.ErrValidation},
		{"unknown column", inTodo, target("archive"), board.ErrInvalidMove},
		{"empty column name", inTodo, target(""), board.ErrInvalidMove},
		{"same column", inTodo, target(board.Todo), board.ErrInvalidMove},
		{"same column doing beats full doing", inDoing, target(board.Doing), board.ErrInvalidMove},
		{"todo to done", inTodo, target(board.Done), board.ErrInvalidMove},
		{"done to todo", inDone, target(board.Todo), board.ErrInvalidMove},
		{"todo into full doing", inTodo, target(board.Doing), board.ErrWIPLimit},
		{"done into full doing", inDone, target(board.Doing), board.ErrWIPLimit},
	}
	before := b.Snapshot()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := b.Move(tc.id, tc.to); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
	after := b.Snapshot()
	for _, c := range []board.Column{board.Todo, board.Doing, board.Done} {
		if !equal(laneIDs(before, c), laneIDs(after, c)) {
			t.Fatalf("refused moves changed %s: %v -> %v", c, laneIDs(before, c), laneIDs(after, c))
		}
	}
}

func TestMove_RightAndLeftLandAtEnd(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 4)
	mustMove(t, b, ids[0], board.Doing)
	c, err := b.Move(ids[1], target(board.Doing))
	if err != nil || c.Column != board.Doing || c.Position != 1 {
		t.Fatalf("right move = %+v, %v", c, err)
	}
	c, err = b.Move(ids[0], target(board.Todo))
	if err != nil || c.Column != board.Todo || c.Position != 2 {
		t.Fatalf("left move = %+v, %v", c, err)
	}
	assertDense(t, b.Snapshot())
}

func TestMove_IntoDoingAtTwoSucceeds(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 3)
	mustMove(t, b, ids[0], board.Doing)
	mustMove(t, b, ids[1], board.Doing)
	mustMove(t, b, ids[2], board.Doing)
	if n := len(b.Snapshot().Lanes[1].Cards); n != board.WIPLimit {
		t.Fatalf("doing holds %d, want %d", n, board.WIPLimit)
	}
}

func TestEdit_ChangesFieldsKeepsPlace(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 3)
	newTitle, newDesc := " New ", "d"
	c, err := b.Edit(ids[1], &newTitle, nil)
	if err != nil || c.Title != "New" || c.Description != "" || c.Column != board.Todo || c.Position != 1 {
		t.Fatalf("edit title = %+v, %v", c, err)
	}
	c, err = b.Edit(ids[1], nil, &newDesc)
	if err != nil || c.Title != "New" || c.Description != "d" || c.Position != 1 {
		t.Fatalf("edit description = %+v, %v", c, err)
	}
	blank := "  "
	if _, err := b.Edit(ids[1], &blank, &newDesc); !errors.Is(err, board.ErrValidation) {
		t.Fatalf("blank title err = %v", err)
	}
	if got := b.Snapshot().Lanes[0].Cards[1].Title; got != "New" {
		t.Fatalf("refused edit changed title to %q", got)
	}
	if _, err := b.Edit("missing", &newTitle, nil); !errors.Is(err, board.ErrNotFound) {
		t.Fatalf("missing err = %v", err)
	}
}

func TestDelete_UnknownIDNotFound(t *testing.T) {
	if err := board.New().Delete("missing"); !errors.Is(err, board.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

// AC-022
func TestDenseOrder_DeleteFirstShiftsUp(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 3)
	if err := b.Delete(ids[0]); err != nil {
		t.Fatal(err)
	}
	s := b.Snapshot()
	if got := laneIDs(s, board.Todo); !equal(got, ids[1:]) {
		t.Fatalf("todo = %v, want %v", got, ids[1:])
	}
	assertDense(t, s)
}

// AC-023
func TestDenseOrder_MoveKeepsBothColumnsDense(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 4)
	mustMove(t, b, ids[3], board.Doing) // todo holds 3, doing holds 1
	mustMove(t, b, ids[0], board.Doing)
	s := b.Snapshot()
	if got := laneIDs(s, board.Todo); !equal(got, ids[1:3]) {
		t.Fatalf("todo = %v", got)
	}
	if got := laneIDs(s, board.Doing); !equal(got, []string{ids[3], ids[0]}) {
		t.Fatalf("doing = %v", got)
	}
	assertDense(t, s)
}

func TestSnapshot_IsACopy(t *testing.T) {
	b := board.New()
	ids := fill(t, b, 2)
	s := b.Snapshot()
	s.Lanes[0].Cards[0].Title = "mutated"
	s.Lanes[0].Cards = s.Lanes[0].Cards[:0]
	mustMove(t, b, ids[0], board.Doing)
	again := b.Snapshot()
	if len(again.Lanes[0].Cards) != 1 || len(again.Lanes[1].Cards) != 1 || again.Lanes[1].Cards[0].Title != "card" {
		t.Fatalf("caller mutation leaked into the board: %+v", again)
	}
}
