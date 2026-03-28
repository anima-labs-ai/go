package anima

import (
	"context"
	"net/url"
	"strconv"
)

// Page represents a single page of results from a paginated API endpoint.
type Page[T any] struct {
	Items      []T              `json:"items"`
	Pagination CursorPagination `json:"pagination"`
}

// CursorPagination contains cursor-based pagination metadata.
type CursorPagination struct {
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

// ListParams are common pagination parameters accepted by list endpoints.
type ListParams struct {
	Cursor string
	Limit  int
}

// ToQuery converts ListParams into URL query values.
func (p ListParams) ToQuery() url.Values {
	q := url.Values{}
	if p.Cursor != "" {
		q.Set("cursor", p.Cursor)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	return q
}

// PageFunc is a function that fetches a single page given a cursor.
type PageFunc[T any] func(ctx context.Context, cursor string) (*Page[T], error)

// ListIterator provides a convenient way to iterate through all pages of a
// paginated API response.
//
//	iter := anima.NewListIterator(fetchPage)
//	for iter.Next(ctx) {
//	    item := iter.Current()
//	    // ...
//	}
//	if err := iter.Err(); err != nil {
//	    // handle error
//	}
type ListIterator[T any] struct {
	fetch   PageFunc[T]
	page    *Page[T]
	index   int
	cursor  string
	err     error
	started bool
	done    bool
}

// NewListIterator creates a new paginated iterator using the given fetch function.
func NewListIterator[T any](fetch PageFunc[T]) *ListIterator[T] {
	return &ListIterator[T]{
		fetch: fetch,
		index: -1,
	}
}

// Next advances the iterator to the next item. It returns true if there is a
// current item available via Current(). When iteration is complete or an error
// occurs, it returns false.
func (it *ListIterator[T]) Next(ctx context.Context) bool {
	if it.done || it.err != nil {
		return false
	}

	// Try to advance within the current page.
	if it.page != nil && it.index+1 < len(it.page.Items) {
		it.index++
		return true
	}

	// If we've already fetched a page and there are no more pages, we're done.
	if it.started && (it.page == nil || !it.page.Pagination.HasMore) {
		it.done = true
		return false
	}

	// Fetch the next page.
	it.started = true
	page, err := it.fetch(ctx, it.cursor)
	if err != nil {
		it.err = err
		return false
	}

	it.page = page
	it.index = 0

	if page.Pagination.NextCursor != nil {
		it.cursor = *page.Pagination.NextCursor
	}

	if len(page.Items) == 0 {
		it.done = true
		return false
	}

	return true
}

// Current returns the current item. Panics if called before Next or after
// iteration is complete.
func (it *ListIterator[T]) Current() T {
	return it.page.Items[it.index]
}

// Err returns the error that caused iteration to stop, if any.
func (it *ListIterator[T]) Err() error {
	return it.err
}

// Page returns the current page of results. Returns nil if no page has been
// fetched yet.
func (it *ListIterator[T]) Page() *Page[T] {
	return it.page
}

// Reset resets the iterator to the beginning.
func (it *ListIterator[T]) Reset() {
	it.page = nil
	it.index = -1
	it.cursor = ""
	it.err = nil
	it.started = false
	it.done = false
}
