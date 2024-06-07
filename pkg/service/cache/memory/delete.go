package memory

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func (s *Service) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}

		s.store.Remove(key)
		if s.store.Has(key) {
			return fmt.Errorf("cannot delete %s", key)
		}
	}

	return nil
}

func (s *Service) DeletePattern(ctx context.Context, pattern string) error {
	expString := "^" + strings.Replace(pattern, "*", ".*", -1)
	r, err := regexp.Compile(expString)
	if err != nil {
		return err
	}

	for _, key := range s.store.Keys() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if r.MatchString(key) {
			s.Delete(ctx, key)
		}
	}

	return nil
}

func (s *Service) Flush(ctx context.Context) (string, error) {
	s.store.Clear()
	s.expirations.Clear()
	return "OK", nil
}
