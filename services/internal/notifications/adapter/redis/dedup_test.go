package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newDedup(t *testing.T) (*Dedup, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	return NewDedup(goredis.NewClient(&goredis.Options{Addr: mr.Addr()}), time.Hour), mr
}

func TestFirstSeenOnlyOnce(t *testing.T) {
	d, _ := newDedup(t)
	ctx := context.Background()

	first, err := d.FirstSeen(ctx, "notify:order.created:7")
	if err != nil {
		t.Fatal(err)
	}
	if !first {
		t.Fatal("first call reported a duplicate")
	}

	again, err := d.FirstSeen(ctx, "notify:order.created:7")
	if err != nil {
		t.Fatal(err)
	}
	if again {
		t.Fatal("second call reported the event as new")
	}
}

func TestForgetLetsEventThroughAgain(t *testing.T) {
	d, _ := newDedup(t)
	ctx := context.Background()

	if _, err := d.FirstSeen(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if err := d.Forget(ctx, "k"); err != nil {
		t.Fatal(err)
	}

	first, err := d.FirstSeen(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if !first {
		t.Fatal("event stayed marked after Forget")
	}
}

func TestMarkExpires(t *testing.T) {
	d, mr := newDedup(t)
	ctx := context.Background()

	if _, err := d.FirstSeen(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	mr.FastForward(2 * time.Hour)

	first, err := d.FirstSeen(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if !first {
		t.Fatal("mark outlived its TTL")
	}
}
