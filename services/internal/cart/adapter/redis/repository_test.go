package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newRepo(t *testing.T) *Repository {
	t.Helper()
	mr := miniredis.RunT(t)
	return New(goredis.NewClient(&goredis.Options{Addr: mr.Addr()}))
}

func TestAddAccumulatesAndReads(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()

	if err := r.Add(ctx, "s1", "hat", 2); err != nil {
		t.Fatal(err)
	}
	if err := r.Add(ctx, "s1", "hat", 3); err != nil {
		t.Fatal(err)
	}
	if err := r.Add(ctx, "s1", "shoe", 1); err != nil {
		t.Fatal(err)
	}

	items, err := r.Items(ctx, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if items["hat"] != 5 || items["shoe"] != 1 {
		t.Fatalf("got %+v, want hat=5 shoe=1", items)
	}
}

func TestSetQtyAndRemove(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()

	_ = r.Add(ctx, "s1", "hat", 9)
	if err := r.SetQty(ctx, "s1", "hat", 2); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(ctx, "s1", "shoe"); err != nil { // удаление отсутствующего — не ошибка
		t.Fatal(err)
	}

	items, _ := r.Items(ctx, "s1")
	if items["hat"] != 2 {
		t.Fatalf("hat = %d, want 2", items["hat"])
	}

	_ = r.Remove(ctx, "s1", "hat")
	items, _ = r.Items(ctx, "s1")
	if len(items) != 0 {
		t.Fatalf("want empty cart, got %+v", items)
	}
}
