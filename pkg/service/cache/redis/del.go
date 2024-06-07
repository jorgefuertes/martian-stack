package redis

import "context"

func (c *Service) Delete(ctx context.Context, keys ...string) error {
	return c.driver.Del(ctx, keys...).Err()
}

func (c *Service) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.driver.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.driver.Del(ctx, keys...).Err()
	}
	return nil
}

func (c *Service) Flush(ctx context.Context) (string, error) {
	return c.driver.FlushAll(ctx).Result()
}
