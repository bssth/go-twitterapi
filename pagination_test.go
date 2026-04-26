package twitterapi

import (
	"context"
	"errors"
	"testing"
)

func TestIterator_HappyPath(t *testing.T) {
	pages := [][]int{{1, 2, 3}, {4, 5}, {6}}
	idx := 0
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		if idx >= len(pages) {
			return Page[int]{}, nil
		}
		p := pages[idx]
		idx++
		return Page[int]{Items: p, NextCursor: "next", HasNextPage: idx < len(pages)}, nil
	}
	it := NewIterator(context.Background(), fetch)
	got := []int{}
	for it.Next() {
		got = append(got, it.Item())
	}
	if it.Err() != nil {
		t.Fatal(it.Err())
	}
	if len(got) != 6 {
		t.Fatalf("got %v", got)
	}
}

func TestIterator_MaxItems(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		return Page[int]{Items: []int{1, 2, 3, 4, 5}, NextCursor: "n", HasNextPage: true}, nil
	}
	it := NewIterator(context.Background(), fetch)
	it.MaxItems = 3
	count := 0
	for it.Next() {
		count++
	}
	if count != 3 {
		t.Fatalf("got %d", count)
	}
}

func TestIterator_MaxPages(t *testing.T) {
	calls := 0
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		calls++
		return Page[int]{Items: []int{1, 2}, NextCursor: "n", HasNextPage: true}, nil
	}
	it := NewIterator(context.Background(), fetch)
	it.MaxPages = 2
	for it.Next() {
	}
	if calls != 2 {
		t.Fatalf("expected 2 pages, fetched %d", calls)
	}
}

func TestIterator_PropagatesError(t *testing.T) {
	want := errors.New("nope")
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		return Page[int]{}, want
	}
	it := NewIterator(context.Background(), fetch)
	if it.Next() {
		t.Fatal("Next should fail")
	}
	if !errors.Is(it.Err(), want) {
		t.Fatalf("err=%v", it.Err())
	}
}

func TestIterator_StopsOnEmptyPage(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		return Page[int]{Items: nil, HasNextPage: false}, nil
	}
	it := NewIterator(context.Background(), fetch)
	if it.Next() {
		t.Fatal("should not advance on empty page")
	}
}

func TestIterator_All(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		return Page[int]{Items: []int{1, 2, 3}, HasNextPage: false}, nil
	}
	it := NewIterator(context.Background(), fetch)
	got, err := it.All(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
}

func TestIterator_StopEarly(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		return Page[int]{Items: []int{1, 2}, HasNextPage: true, NextCursor: "x"}, nil
	}
	it := NewIterator(context.Background(), fetch)
	it.Next()
	it.Stop()
	if it.Next() {
		t.Fatal("should be stopped")
	}
}

func TestIterator_CursorTracking(t *testing.T) {
	page := 0
	fetch := func(ctx context.Context, cursor string) (Page[int], error) {
		page++
		if page == 1 {
			if cursor != "" {
				t.Fatalf("first page cursor %q", cursor)
			}
			return Page[int]{Items: []int{1}, NextCursor: "abc", HasNextPage: true}, nil
		}
		if cursor != "abc" {
			t.Fatalf("second page cursor %q", cursor)
		}
		return Page[int]{Items: []int{2}, HasNextPage: false}, nil
	}
	it := NewIterator(context.Background(), fetch)
	for it.Next() {
	}
	if it.Cursor() != "" {
		// After last page, cursor was consumed.
		t.Logf("final cursor=%q (acceptable)", it.Cursor())
	}
}
