package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// New создаёт redis-клиент и один раз пингует его. Close — на стороне вызывающего.
func New(ctx context.Context, addr string, db int) (*goredis.Client, error) {
	c := goredis.NewClient(&goredis.Options{Addr: addr, DB: db})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := c.Ping(pingCtx).Err(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return c, nil
}
