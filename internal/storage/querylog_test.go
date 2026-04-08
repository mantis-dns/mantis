package storage

import (
	"context"
	"testing"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

func openTestDB(t *testing.T) *pebble.DB {
	t.Helper()
	db, err := pebble.Open(t.TempDir(), &pebble.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestQueryLogAppendAndQuery(t *testing.T) {
	db := openTestDB(t)
	store := NewQueryLogStore(db)
	defer store.Close()

	ctx := context.Background()
	base := time.Now()

	for i := range 100 {
		entry := &domain.QueryLogEntry{
			Timestamp: base.Add(time.Duration(i) * time.Second),
			ClientIP:  "192.168.1.10",
			Domain:    "example.com",
			QueryType: domain.QTypeA,
			Result:    "allowed",
			LatencyUs: 500,
		}
		if err := store.Append(ctx, entry); err != nil {
			t.Fatalf("Append failed: %v", err)
		}
	}

	entries, total, err := store.Query(ctx, domain.QueryLogFilter{
		PerPage: 20,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if total != 100 {
		t.Errorf("expected total 100, got %d", total)
	}
	if len(entries) != 20 {
		t.Errorf("expected 20 entries, got %d", len(entries))
	}
}

func TestQueryLogTimeRangeQuery(t *testing.T) {
	db := openTestDB(t)
	store := NewQueryLogStore(db)
	defer store.Close()

	ctx := context.Background()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range 10 {
		entry := &domain.QueryLogEntry{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			ClientIP:  "10.0.0.1",
			Domain:    "test.com",
			QueryType: domain.QTypeA,
			Result:    "allowed",
			LatencyUs: 100,
		}
		if err := store.Append(ctx, entry); err != nil {
			t.Fatal(err)
		}
	}

	entries, total, err := store.Query(ctx, domain.QueryLogFilter{
		From:    base.Add(2 * time.Hour),
		To:      base.Add(5 * time.Hour),
		PerPage: 50,
		Page:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	// From=2h, To=5h: entries at hours 2, 3, 4, 5 = 4 entries (both bounds inclusive).
	if total != 4 {
		t.Errorf("expected 4 entries in range, got %d", total)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
}

func TestQueryLogDeleteBefore(t *testing.T) {
	db := openTestDB(t)
	store := NewQueryLogStore(db)
	defer store.Close()

	ctx := context.Background()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range 10 {
		entry := &domain.QueryLogEntry{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			ClientIP:  "10.0.0.1",
			Domain:    "test.com",
			QueryType: domain.QTypeA,
			Result:    "allowed",
			LatencyUs: 100,
		}
		store.Append(ctx, entry)
	}

	deleted, err := store.DeleteBefore(ctx, base.Add(5*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 5 {
		t.Errorf("expected 5 deleted, got %d", deleted)
	}

	_, total, _ := store.Query(ctx, domain.QueryLogFilter{PerPage: 100, Page: 1})
	if total != 5 {
		t.Errorf("expected 5 remaining, got %d", total)
	}
}

func TestQueryLogFilterByDomain(t *testing.T) {
	db := openTestDB(t)
	store := NewQueryLogStore(db)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	store.Append(ctx, &domain.QueryLogEntry{Timestamp: now, Domain: "foo.com", ClientIP: "1.1.1.1", Result: "allowed"})
	store.Append(ctx, &domain.QueryLogEntry{Timestamp: now.Add(time.Second), Domain: "bar.com", ClientIP: "1.1.1.1", Result: "blocked"})
	store.Append(ctx, &domain.QueryLogEntry{Timestamp: now.Add(2 * time.Second), Domain: "foo.com", ClientIP: "2.2.2.2", Result: "allowed"})

	entries, total, err := store.Query(ctx, domain.QueryLogFilter{Domain: "foo.com", PerPage: 50, Page: 1})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	for _, e := range entries {
		if e.Domain != "foo.com" {
			t.Errorf("unexpected domain %s", e.Domain)
		}
	}
}
