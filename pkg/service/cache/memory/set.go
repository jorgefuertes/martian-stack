package memory

import (
	"context"
	"time"
)

func (s *Service) Set(ctx context.Context, key string, v any, expire time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.store.Set(key, newValue(v))
	if expire == 0 {
		return nil
	}

	expireAt := time.Now().Add(expire)
	s.expirations.Set(key, expireAt)

	return nil
}
