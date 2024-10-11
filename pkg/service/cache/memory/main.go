package memory

import (
	"sync"
	"time"
)

type Service struct {
	store       map[string][]byte
	expirations map[string]time.Time
	lock        *sync.Mutex
}

func New() *Service {
	s := &Service{
		store:       make(map[string][]byte),
		expirations: make(map[string]time.Time),
		lock:        &sync.Mutex{},
	}

	s.StartExpirationController()

	return s
}

func (s *Service) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store = nil
	s.expirations = nil

	return nil
}
