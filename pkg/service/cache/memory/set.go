package memory

import (
	"context"
	"encoding/json"
	"time"
)

func encode(v any) ([]byte, error) {
	if b, ok := v.([]byte); ok {
		return b, nil
	}

	return json.Marshal(v)
}

func (s *Service) Set(ctx context.Context, key string, v any, expire time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b, err := encode(v)
	if err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[key] = b
	if expire == 0 {
		return nil
	}

	expireAt := time.Now().Add(expire)
	s.expirations[key] = expireAt

	return nil
}
