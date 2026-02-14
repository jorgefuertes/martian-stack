package cache

import (
	"context"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/service/cache/memory"
	"github.com/jorgefuertes/martian-stack/pkg/service/cache/redis"
)

type Service interface {
	Close() error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetFloat(ctx context.Context, key string) (float64, error)
	GetBytes(ctx context.Context, key string) ([]byte, error)
	Exists(ctx context.Context, key string) bool
	Keys(ctx context.Context, pattern string) ([]string, error)
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Flush(ctx context.Context) (string, error)
}

var (
	_ Service = &redis.Service{}
	_ Service = &memory.Service{}
)
