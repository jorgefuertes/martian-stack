package memory

import (
	"context"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type Service struct {
	store       cmap.ConcurrentMap[string, value]
	expirations cmap.ConcurrentMap[string, time.Time]
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewService() *Service {
	s := &Service{store: cmap.New[value](), expirations: cmap.New[time.Time]()}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.StartExpirationController()

	return s
}

func (s *Service) Close() error {
	s.store.Clear()
	s.expirations.Clear()
	s.cancel()

	return nil
}
