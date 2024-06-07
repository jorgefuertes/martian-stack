package redis

import (
	"fmt"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"

	driver "github.com/redis/go-redis/v9"
)

type Service struct {
	driver *driver.Client
	log    *logger.Service
}

func NewService(log *logger.Service, host string, port int, user, pass string, db int) *Service {
	return &Service{
		log: log,
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
