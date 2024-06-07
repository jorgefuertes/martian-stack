package memory

import (
	"context"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	v, ok := s.store.Get(key)
	if ok {
		return v.String(), nil
	}

	return "", ErrKeyNotFound
}

func (s *Service) GetInt(ctx context.Context, key string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	v, ok := s.store.Get(key)
	if ok {
		return v.int()
	}

	return 0, ErrKeyNotFound
}

func (s *Service) GetFloat(ctx context.Context, key string) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	v, ok := s.store.Get(key)
	if ok {
		return v.float64()
	}

	return 0, ErrKeyNotFound
}

func (s *Service) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return []byte{}, err
	}

	v, ok := s.store.Get(key)
	if ok {
		return v.bytes(), nil
	}

	return []byte{}, ErrKeyNotFound
}

func (s *Service) GetObjectID(ctx context.Context, key string) (primitive.ObjectID, error) {
	if err := ctx.Err(); err != nil {
		return primitive.NilObjectID, err
	}

	v, ok := s.store.Get(key)
	if ok {
		return primitive.ObjectIDFromHex(v.String())
	}

	return primitive.NilObjectID, ErrKeyNotFound
}

func (s *Service) Exists(ctx context.Context, key string) bool {
	if err := ctx.Err(); err != nil {
		return false
	}

	return s.store.Has(key)
}

func (s *Service) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys := make([]string, 0)
	expString := "^" + strings.Replace(pattern, "*", ".*", -1)
	r, err := regexp.Compile(expString)
	if err != nil {
		return keys, err
	}

	for _, key := range s.store.Keys() {
		if err := ctx.Err(); err != nil {
			return keys, err
		}
		if r.MatchString(key) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}
