package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Корзина живёт неделю с момента последнего изменения.
const cartTTL = 7 * 24 * time.Hour

// Repository хранит корзину как redis-хеш cart:{session}: поле = uuid товара, значение = количество.
type Repository struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *Repository {
	return &Repository{rdb: rdb}
}

func (r *Repository) Add(ctx context.Context, session, uuid string, delta int32) error {
	k := key(session)
	if err := r.rdb.HIncrBy(ctx, k, uuid, int64(delta)).Err(); err != nil {
		return fmt.Errorf("hincrby: %w", err)
	}
	return r.touch(ctx, k)
}

func (r *Repository) SetQty(ctx context.Context, session, uuid string, qty int32) error {
	k := key(session)
	if err := r.rdb.HSet(ctx, k, uuid, qty).Err(); err != nil {
		return fmt.Errorf("hset: %w", err)
	}
	return r.touch(ctx, k)
}

func (r *Repository) Remove(ctx context.Context, session, uuid string) error {
	if err := r.rdb.HDel(ctx, key(session), uuid).Err(); err != nil {
		return fmt.Errorf("hdel: %w", err)
	}
	return nil
}

func (r *Repository) Items(ctx context.Context, session string) (map[string]int32, error) {
	raw, err := r.rdb.HGetAll(ctx, key(session)).Result()
	if err != nil {
		return nil, fmt.Errorf("hgetall: %w", err)
	}

	out := make(map[string]int32, len(raw))
	for uuid, v := range raw {
		var qty int32
		if _, err := fmt.Sscan(v, &qty); err != nil {
			return nil, fmt.Errorf("parse qty for %s: %w", uuid, err)
		}
		out[uuid] = qty
	}
	return out, nil
}

func (r *Repository) touch(ctx context.Context, k string) error {
	if err := r.rdb.Expire(ctx, k, cartTTL).Err(); err != nil {
		return fmt.Errorf("expire: %w", err)
	}
	return nil
}

func key(session string) string {
	return "cart:" + session
}
