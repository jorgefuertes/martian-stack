package redis

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

func (c *Service) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	b, err := encode(value)
	if err != nil {
		return err
	}

	res := c.driver.Set(ctx, key, b, expiration)
	return res.Err()
}
