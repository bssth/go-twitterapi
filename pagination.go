package twitterapi

import "context"

// Page is a single page from any cursor-paginated endpoint.
type Page[T any] struct {
	Items       []T
	NextCursor  string
	HasNextPage bool
}

// PageFunc fetches one page given a cursor. Implementations are returned by
// service helpers like Users.FollowersPage.
type PageFunc[T any] func(ctx context.Context, cursor string) (Page[T], error)

// Iterator walks every page produced by a PageFunc.
//
//	it := client.Users.Followers(ctx, "elonmusk", nil)
//	for it.Next() {
//	    fmt.Println(it.Item().UserName)
//	}
//	if err := it.Err(); err != nil { ... }
//
// Iterators stop at the first error or when HasNextPage becomes false.
type Iterator[T any] struct {
	ctx   context.Context
	fetch PageFunc[T]

	cursor    string
	started   bool
	hasNext   bool
	stopped   bool
	buf       []T
	idx       int
	cur       T
	err       error
	pageCount int
	itemCount int

	// MaxPages bounds the iterator (0 = unlimited).
	MaxPages int
	// MaxItems bounds the iterator (0 = unlimited).
	MaxItems int
}

// NewIterator wires a PageFunc into a stateful Iterator.
func NewIterator[T any](ctx context.Context, fetch PageFunc[T]) *Iterator[T] {
	return &Iterator[T]{ctx: ctx, fetch: fetch, hasNext: true}
}

// Next advances to the next item, fetching pages on demand. Returns false on
// exhaustion or error (check Err).
func (it *Iterator[T]) Next() bool {
	if it.stopped || it.err != nil {
		return false
	}
	if it.MaxItems > 0 && it.itemCount >= it.MaxItems {
		return false
	}
	if it.idx < len(it.buf) {
		it.cur = it.buf[it.idx]
		it.idx++
		it.itemCount++
		return true
	}
	if !it.hasNext {
		return false
	}
	if it.MaxPages > 0 && it.pageCount >= it.MaxPages {
		return false
	}
	page, err := it.fetch(it.ctx, it.cursor)
	if err != nil {
		it.err = err
		it.stopped = true
		return false
	}
	it.pageCount++
	it.buf = page.Items
	it.idx = 0
	it.cursor = page.NextCursor
	it.hasNext = page.HasNextPage && page.NextCursor != ""
	it.started = true
	if len(it.buf) == 0 {
		// Empty page but server says there's more — guard against infinite loops.
		if it.hasNext {
			return it.Next()
		}
		return false
	}
	it.cur = it.buf[0]
	it.idx = 1
	it.itemCount++
	return true
}

// Item returns the current value. Only valid between a Next() that returned
// true and the subsequent Next/Stop call.
func (it *Iterator[T]) Item() T { return it.cur }

// Err returns the first error, if any.
func (it *Iterator[T]) Err() error { return it.err }

// Stop terminates the iterator early.
func (it *Iterator[T]) Stop() { it.stopped = true }

// Cursor returns the cursor that would be passed to the next fetch (the
// position after the most-recent page).
func (it *Iterator[T]) Cursor() string { return it.cursor }

// All collects up to limit items (or every item when limit <= 0). Convenience
// for small result sets — beware of cost on large ones.
func (it *Iterator[T]) All(limit int) ([]T, error) {
	if limit > 0 {
		it.MaxItems = limit
	}
	out := make([]T, 0, 64)
	for it.Next() {
		out = append(out, it.Item())
	}
	return out, it.Err()
}
