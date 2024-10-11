package memory

import (
	"context"
	"regexp"
	"strings"
	"time"
)

func (s *Service) Delete(ctx context.Context, keys ...string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}

		delete(s.expirations, key)
		delete(s.store, key)
	}

	return nil
}

func (s *Service) DeletePattern(ctx context.Context, pattern string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	expString := "^" + strings.Replace(pattern, "*", ".*", -1)
	r, err := regexp.Compile(expString)
	if err != nil {
		return err
	}

	for k := range s.store {
		if err := ctx.Err(); err != nil {
			return err
		}
		if r.MatchString(k) {
			delete(s.store, k)
		}
	}

	return nil
}

func (s *Service) Flush(ctx context.Context) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.expirations = make(map[string]time.Time)
	s.store = make(map[string][]byte)

	return "OK", nil
}
