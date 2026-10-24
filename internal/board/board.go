// Package board holds the Board aggregate (AGG-001): the three fixed columns,
// the cards on them, and the rules INV-001..INV-005.
//
// The Board owns its sync.Mutex (ADR-0006): every exported method locks it for
// its whole body, so the checks and the state change of one command form one
// critical section, and Snapshot returns a copy. The package knows nothing about
// HTTP (ADR-0004/R1).
package board

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// Column is one of the three fixed board lanes (INV-005).
type Column string

// The three columns, in board order.
const (
	Todo  Column = "todo"
	Doing Column = "doing"
	Done  Column = "done"
)

// WIPLimit is the maximum number of cards the doing column holds (INV-002).
const WIPLimit = 3

// columns lists the board's columns in their fixed order (INV-005); a column's
// index here is its index in Board.lanes and Snapshot.Lanes.
var columns = [...]Column{Todo, Doing, Done}

// Refusals. Each maps onto exactly one CON-004 error code.
var (
	ErrNotFound    = errors.New("card not found")
	ErrValidation  = errors.New("validation failed")
	ErrInvalidMove = errors.New("invalid move")
	ErrWIPLimit    = errors.New("wip limit exceeded")
)

// Card is a value copy of a card as it sits on the board.
type Card struct {
	ID          string
	Title       string
	Description string
	Column      Column
	Position    int
}

// Lane is a value copy of one column and its cards, ordered by position.
type Lane struct {
	Name  Column
	Cards []Card
}

// Snapshot is a copy of the whole board: always three lanes in the order
// todo, doing, done. Mutating it never affects the Board.
type Snapshot struct {
	Lanes [len(columns)]Lane
}

type card struct {
	id          string
	title       string
	description string
}

// Board is the single shared board. The zero value is not usable; call New.
type Board struct {
	mu     sync.Mutex
	lanes  [len(columns)][]card // index = column index; slice index = position (dense by construction, INV-004)
	nextID int
}

// New returns an empty board.
func New() *Board {
	return &Board{}
}

// Snapshot returns a copy of the board.
func (b *Board) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	var s Snapshot
	for ci, lane := range b.lanes {
		cards := make([]Card, len(lane))
		for pos := range lane {
			cards[pos] = b.cardAt(ci, pos)
		}
		s.Lanes[ci] = Lane{Name: columns[ci], Cards: cards}
	}
	return s
}

// Add puts a new card at the end of todo. The title is trimmed and must not be
// blank (INV-001); the description is optional and stored as given.
func (b *Board) Add(title, description string) (Card, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	title, err := validTitle(title)
	if err != nil {
		return Card{}, err
	}
	b.nextID++
	c := card{id: strconv.Itoa(b.nextID), title: title, description: description}
	ti := columnIndex(Todo)
	b.lanes[ti] = append(b.lanes[ti], c)
	return b.cardAt(ti, len(b.lanes[ti])-1), nil
}

// Move moves a card one column left or right; it lands at the end of the
// target (INV-003, INV-004). A nil to means the request named no readable
// target (missing, non-string or malformed); any non-nil to is a named target,
// whatever its value, the empty string included. The first failing check
// wins, in the ADR-0007 order: ErrNotFound, ErrValidation (to is nil),
// ErrInvalidMove (unknown, empty, same or non-neighbouring column),
// ErrWIPLimit (doing already holds WIPLimit cards, INV-002). All checks and
// the change run in one critical section; a refused move changes nothing.
func (b *Board) Move(id string, to *Column) (Card, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	from, pos, ok := b.find(id)
	if !ok {
		return Card{}, ErrNotFound
	}
	if to == nil {
		return Card{}, ErrValidation
	}
	ti := columnIndex(*to)
	if ti < 0 || (ti != from-1 && ti != from+1) {
		return Card{}, ErrInvalidMove
	}
	if *to == Doing && len(b.lanes[ti]) >= WIPLimit {
		return Card{}, ErrWIPLimit
	}

	c := b.lanes[from][pos]
	b.lanes[from] = slices.Delete(b.lanes[from], pos, pos+1)
	b.lanes[ti] = append(b.lanes[ti], c)
	return b.cardAt(ti, len(b.lanes[ti])-1), nil
}

// Edit changes the title and/or the description of a card; a nil argument
// leaves that field as it is. The card keeps its column and position. A new
// title is trimmed and must not be blank (INV-001). A refused edit changes
// nothing.
func (b *Board) Edit(id string, title, description *string) (Card, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ci, pos, ok := b.find(id)
	if !ok {
		return Card{}, ErrNotFound
	}
	c := b.lanes[ci][pos]
	if title != nil {
		t, err := validTitle(*title)
		if err != nil {
			return Card{}, err
		}
		c.title = t
	}
	if description != nil {
		c.description = *description
	}
	b.lanes[ci][pos] = c
	return b.cardAt(ci, pos), nil
}

// Delete removes a card; the cards after it in its column shift up by one
// (INV-004).
func (b *Board) Delete(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	ci, pos, ok := b.find(id)
	if !ok {
		return ErrNotFound
	}
	b.lanes[ci] = slices.Delete(b.lanes[ci], pos, pos+1)
	return nil
}

// Helpers below are lock-free and must be called with b.mu held
// (sync.Mutex is not reentrant, ADR-0006).

func (b *Board) find(id string) (ci, pos int, ok bool) {
	for ci, lane := range b.lanes {
		for pos, c := range lane {
			if c.id == id {
				return ci, pos, true
			}
		}
	}
	return 0, 0, false
}

func (b *Board) cardAt(ci, pos int) Card {
	c := b.lanes[ci][pos]
	return Card{ID: c.id, Title: c.title, Description: c.description, Column: columns[ci], Position: pos}
}

func columnIndex(c Column) int {
	for i, name := range columns {
		if name == c {
			return i
		}
	}
	return -1
}

func validTitle(title string) (string, error) {
	t := strings.TrimSpace(title)
	if t == "" {
		return "", ErrValidation
	}
	return t, nil
}
