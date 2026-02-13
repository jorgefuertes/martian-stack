package redis

import (
	"fmt"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"

	driver "github.com/redis/go-redis/v9"
)

const (
	DefaultHost = "localhost"
	DefaultPort = 6379
	DefaultDB   = 0
)

type Service struct {
	driver *driver.Client
}

func New(log *logger.Service, host string, port int, user, pass string, db int) *Service {
	return &Service{
		driver: driver.NewClient(&driver.Options{
			Addr:                  fmt.Sprintf("%s:%d", host, port),
			DB:                    db,
			Username:              user,
			Password:              pass,
			ContextTimeoutEnabled: true,
		}),
	}
}

func (c *Service) Close() error {
	return c.driver.Close()
}
