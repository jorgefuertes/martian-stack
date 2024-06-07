package redis

import (
	"context"
	"time"
)

func (c *Service) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	res := c.driver.Set(ctx, key, value, expiration)
	return res.Err()
}
