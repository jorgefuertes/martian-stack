package redis

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) Get(ctx context.Context, key string, dest any) error {
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return res.Err()
	}

	b, err := res.Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dest)
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
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return []byte{}, res.Err()
	}

	return res.Bytes()
}

func (s *Service) GetObjectID(ctx context.Context, key string) (primitive.ObjectID, error) {
	var v primitive.ObjectID
	err := s.Get(ctx, key, &v)
	return v, err
}

func (s *Service) Exists(ctx context.Context, key string) bool {
	res, err := s.driver.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	return res == 1
}

func (s *Service) Keys(ctx context.Context, pattern string) ([]string, error) {
	return s.driver.Keys(ctx, pattern).Result()
}
