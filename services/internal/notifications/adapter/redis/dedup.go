package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Dedup помнит обработанные события. TTL, а не вечная запись: повторная доставка
// приходит в пределах ретраев брокера, а не через неделю, и ключи не копятся.
type Dedup struct {
	rdb *goredis.Client
	ttl time.Duration
}

func NewDedup(rdb *goredis.Client, ttl time.Duration) *Dedup {
	return &Dedup{rdb: rdb, ttl: ttl}
}

// FirstSeen атомарно ставит отметку и говорит, поставил ли её именно этот вызов.
func (d *Dedup) FirstSeen(ctx context.Context, key string) (bool, error) {
	ok, err := d.rdb.SetNX(ctx, key, time.Now().UTC().Format(time.RFC3339), d.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("setnx %s: %w", key, err)
	}
	return ok, nil
}

// Forget снимает отметку, чтобы повторная доставка события снова дошла до почты.
func (d *Dedup) Forget(ctx context.Context, key string) error {
	if err := d.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("del %s: %w", key, err)
	}
	return nil
}
