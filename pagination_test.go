package anima

import (
	"context"
	"fmt"
	"testing"
)

func TestListIterator_SinglePage(t *testing.T) {
	cursor2 := ""
	fetch := func(ctx context.Context, cursor string) (*Page[string], error) {
		_ = cursor2
		return &Page[string]{
			Items: []string{"a", "b", "c"},
			Pagination: CursorPagination{
				HasMore: false,
			},
		}, nil
	}

	iter := NewListIterator(fetch)
	var items []string
	for iter.Next(context.Background()) {
		items = append(items, iter.Current())
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
	if items[0] != "a" || items[1] != "b" || items[2] != "c" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestListIterator_MultiplePages(t *testing.T) {
	pages := map[string]*Page[int]{
		"": {
			Items: []int{1, 2},
			Pagination: CursorPagination{
				NextCursor: strPtr("page2"),
				HasMore:    true,
			},
		},
		"page2": {
			Items: []int{3, 4},
			Pagination: CursorPagination{
				NextCursor: strPtr("page3"),
				HasMore:    true,
			},
		},
		"page3": {
			Items: []int{5},
			Pagination: CursorPagination{
				HasMore: false,
			},
		},
	}

	fetch := func(ctx context.Context, cursor string) (*Page[int], error) {
		page, ok := pages[cursor]
		if !ok {
			return nil, fmt.Errorf("unexpected cursor: %s", cursor)
		}
		return page, nil
	}

	iter := NewListIterator(fetch)
	var items []int
	for iter.Next(context.Background()) {
		items = append(items, iter.Current())
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d: %v", len(items), items)
	}
	for i, want := range []int{1, 2, 3, 4, 5} {
		if items[i] != want {
			t.Errorf("items[%d] = %d, want %d", i, items[i], want)
		}
	}
}

func TestListIterator_EmptyPage(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) (*Page[string], error) {
		return &Page[string]{
			Items:      []string{},
			Pagination: CursorPagination{HasMore: false},
		}, nil
	}

	iter := NewListIterator(fetch)
	count := 0
	for iter.Next(context.Background()) {
		count++
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 iterations, got %d", count)
	}
}

func TestListIterator_FetchError(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) (*Page[string], error) {
		return nil, fmt.Errorf("network failure")
	}

	iter := NewListIterator(fetch)
	if iter.Next(context.Background()) {
		t.Error("Next should return false on error")
	}
	if iter.Err() == nil {
		t.Error("Err() should return the error")
	}
}

func TestListIterator_Reset(t *testing.T) {
	callCount := 0
	fetch := func(ctx context.Context, cursor string) (*Page[string], error) {
		callCount++
		return &Page[string]{
			Items:      []string{"x"},
			Pagination: CursorPagination{HasMore: false},
		}, nil
	}

	iter := NewListIterator(fetch)

	// First iteration.
	for iter.Next(context.Background()) {
	}
	if callCount != 1 {
		t.Errorf("expected 1 fetch, got %d", callCount)
	}

	// Reset and iterate again.
	iter.Reset()
	for iter.Next(context.Background()) {
	}
	if callCount != 2 {
		t.Errorf("expected 2 fetches after reset, got %d", callCount)
	}
}

func TestListIterator_Page(t *testing.T) {
	page := &Page[string]{
		Items:      []string{"a"},
		Pagination: CursorPagination{HasMore: false},
	}
	fetch := func(ctx context.Context, cursor string) (*Page[string], error) {
		return page, nil
	}

	iter := NewListIterator(fetch)
	if iter.Page() != nil {
		t.Error("Page() should be nil before first Next")
	}

	iter.Next(context.Background())
	if iter.Page() == nil {
		t.Error("Page() should not be nil after Next")
	}
	if len(iter.Page().Items) != 1 {
		t.Errorf("expected 1 item in page, got %d", len(iter.Page().Items))
	}
}

func TestListParams_ToQuery(t *testing.T) {
	p := ListParams{Cursor: "abc123", Limit: 25}
	q := p.ToQuery()
	if q.Get("cursor") != "abc123" {
		t.Errorf("cursor = %q, want abc123", q.Get("cursor"))
	}
	if q.Get("limit") != "25" {
		t.Errorf("limit = %q, want 25", q.Get("limit"))
	}

	empty := ListParams{}
	q2 := empty.ToQuery()
	if q2.Get("cursor") != "" || q2.Get("limit") != "" {
		t.Error("empty ListParams should produce empty query")
	}
}

func strPtr(s string) *string { return &s }
