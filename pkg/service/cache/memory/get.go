package memory

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) Get(ctx context.Context, key string, dest any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	v, ok := s.store[key]
	if !ok {
		return ErrKeyNotFound
	}

	return json.Unmarshal(v, dest)
}

func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	var v string
	err := s.Get(ctx, key, &v)
	return v, err
}

func (s *Service) GetInt(ctx context.Context, key string) (int, error) {
	var v int
	err := s.Get(ctx, key, &v)
	return v, err
}

func (s *Service) GetFloat(ctx context.Context, key string) (float64, error) {
	var v float64
	err := s.Get(ctx, key, &v)
	return v, err
}

func (s *Service) GetBytes(ctx context.Context, key string) ([]byte, error) {
	v, ok := s.store[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return v, nil
}

func (s *Service) GetObjectID(ctx context.Context, key string) (primitive.ObjectID, error) {
	var v primitive.ObjectID
	err := s.Get(ctx, key, &v)
	return v, err
}

func (s *Service) Exists(ctx context.Context, key string) bool {
	if err := ctx.Err(); err != nil {
		return false
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	for k := range s.store {
		if k == key {
			return true
		}
	}

	return false
}

func (s *Service) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys := make([]string, 0)
	expString := "^" + strings.Replace(pattern, "*", ".*", -1)
	r, err := regexp.Compile(expString)
	if err != nil {
		return keys, err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	for k := range s.store {
		if err := ctx.Err(); err != nil {
			return keys, err
		}
		if r.MatchString(k) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}
