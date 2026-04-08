package storage

import (
	"context"
	"testing"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
)

func TestLeaseCRUD(t *testing.T) {
	db := openTestDB(t)
	store := NewLeaseStore(db)
	ctx := context.Background()

	lease := &domain.DhcpLease{
		MAC:      "AA:BB:CC:DD:EE:FF",
		IP:       "192.168.1.100",
		Hostname: "desktop",
		LeaseEnd: time.Now().Add(24 * time.Hour),
	}

	if err := store.Create(ctx, lease); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := store.Get(ctx, "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.IP != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", got.IP)
	}
	if got.Hostname != "desktop" {
		t.Errorf("expected hostname desktop, got %s", got.Hostname)
	}
}

func TestLeaseGetByIP(t *testing.T) {
	db := openTestDB(t)
	store := NewLeaseStore(db)
	ctx := context.Background()

	lease := &domain.DhcpLease{
		MAC:      "11:22:33:44:55:66",
		IP:       "10.0.0.50",
		LeaseEnd: time.Now().Add(time.Hour),
	}
	store.Create(ctx, lease)

	got, err := store.GetByIP(ctx, "10.0.0.50")
	if err != nil {
		t.Fatalf("GetByIP failed: %v", err)
	}
	if got.MAC != "11:22:33:44:55:66" {
		t.Errorf("expected MAC 11:22:33:44:55:66, got %s", got.MAC)
	}
}

func TestLeaseList(t *testing.T) {
	db := openTestDB(t)
	store := NewLeaseStore(db)
	ctx := context.Background()

	for i := range 3 {
		store.Create(ctx, &domain.DhcpLease{
			MAC:      "AA:BB:CC:DD:EE:" + string(rune('A'+i)),
			IP:       "192.168.1." + string(rune('1'+i)),
			LeaseEnd: time.Now().Add(time.Hour),
		})
	}

	leases, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(leases) != 3 {
		t.Errorf("expected 3, got %d", len(leases))
	}
}

func TestLeaseDelete(t *testing.T) {
	db := openTestDB(t)
	store := NewLeaseStore(db)
	ctx := context.Background()

	store.Create(ctx, &domain.DhcpLease{
		MAC:      "DE:AD:BE:EF:00:01",
		IP:       "10.0.0.99",
		LeaseEnd: time.Now().Add(time.Hour),
	})

	if err := store.Delete(ctx, "DE:AD:BE:EF:00:01"); err != nil {
		t.Fatal(err)
	}

	_, err := store.Get(ctx, "DE:AD:BE:EF:00:01")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	_, err = store.GetByIP(ctx, "10.0.0.99")
	if err != domain.ErrNotFound {
		t.Errorf("expected IP cross-index cleaned, got %v", err)
	}
}

func TestLeaseDeleteExpired(t *testing.T) {
	db := openTestDB(t)
	store := NewLeaseStore(db)
	ctx := context.Background()

	store.Create(ctx, &domain.DhcpLease{
		MAC:      "AA:00:00:00:00:01",
		IP:       "10.0.0.1",
		LeaseEnd: time.Now().Add(-time.Hour),
	})
	store.Create(ctx, &domain.DhcpLease{
		MAC:      "AA:00:00:00:00:02",
		IP:       "10.0.0.2",
		LeaseEnd: time.Now().Add(time.Hour),
	})
	store.Create(ctx, &domain.DhcpLease{
		MAC:      "AA:00:00:00:00:03",
		IP:       "10.0.0.3",
		LeaseEnd: time.Now().Add(-2 * time.Hour),
		IsStatic: true,
	})

	deleted, err := store.DeleteExpired(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 expired deleted, got %d", deleted)
	}

	leases, _ := store.List(ctx)
	if len(leases) != 2 {
		t.Errorf("expected 2 remaining, got %d", len(leases))
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	db := openTestDB(t)
	store := NewSettingsStore(db)
	ctx := context.Background()

	if err := store.Set(ctx, "dns.listen", "0.0.0.0:53"); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(ctx, "dns.cache_size", "10000"); err != nil {
		t.Fatal(err)
	}

	val, err := store.Get(ctx, "dns.listen")
	if err != nil {
		t.Fatal(err)
	}
	if val != "0.0.0.0:53" {
		t.Errorf("expected 0.0.0.0:53, got %s", val)
	}

	all, err := store.GetAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 settings, got %d", len(all))
	}
}
