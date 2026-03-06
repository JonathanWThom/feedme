package api

import (
	"testing"
	"time"
)

func TestCachedSource_StoreAndFetchItem(t *testing.T) {
	cs := NewCachedSource(500 * time.Millisecond)

	item := &Item{ID: 42, Title: "Test Story"}
	cs.StoreItems([]*Item{item})

	got, err := cs.FetchItem(1)
	if err != nil {
		t.Fatalf("FetchItem(1) unexpected error: %v", err)
	}
	if got.Title != "Test Story" {
		t.Errorf("FetchItem(1).Title = %q, want %q", got.Title, "Test Story")
	}
}

func TestCachedSource_FetchItemNotFound(t *testing.T) {
	cs := NewCachedSource(500 * time.Millisecond)

	_, err := cs.FetchItem(999)
	if err == nil {
		t.Error("FetchItem(999) expected error, got nil")
	}
}

func TestCachedSource_FetchItems(t *testing.T) {
	cs := NewCachedSource(500 * time.Millisecond)

	items := []*Item{
		{ID: 1, Title: "First"},
		{ID: 2, Title: "Second"},
		{ID: 3, Title: "Third"},
	}
	cs.StoreItems(items)

	got, err := cs.FetchItems([]int{1, 3})
	if err != nil {
		t.Fatalf("FetchItems unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("FetchItems returned %d items, want 2", len(got))
	}
	if got[0].Title != "First" {
		t.Errorf("got[0].Title = %q, want %q", got[0].Title, "First")
	}
	if got[1].Title != "Third" {
		t.Errorf("got[1].Title = %q, want %q", got[1].Title, "Third")
	}
}

func TestCachedSource_StoreItemsReturnsIDs(t *testing.T) {
	cs := NewCachedSource(500 * time.Millisecond)

	items := []*Item{
		{Title: "A"},
		{Title: "B"},
	}
	ids := cs.StoreItems(items)

	if len(ids) != 2 {
		t.Fatalf("StoreItems returned %d ids, want 2", len(ids))
	}
	if ids[0] != 1 || ids[1] != 2 {
		t.Errorf("StoreItems ids = %v, want [1 2]", ids)
	}
}

func TestCachedSource_StoreItemsClearsOldCache(t *testing.T) {
	cs := NewCachedSource(500 * time.Millisecond)

	cs.StoreItems([]*Item{{Title: "Old"}})
	cs.StoreItems([]*Item{{Title: "New"}})

	got, err := cs.FetchItem(1)
	if err != nil {
		t.Fatalf("FetchItem(1) unexpected error: %v", err)
	}
	if got.Title != "New" {
		t.Errorf("FetchItem(1).Title = %q, want %q (old cache not cleared)", got.Title, "New")
	}
}
