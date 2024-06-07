package redis

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return "", res.Err()
	}

	return res.Result()
}

func (s *Service) GetInt(ctx context.Context, key string) (int, error) {
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return 0, res.Err()
	}

	return res.Int()
}

func (s *Service) GetFloat(ctx context.Context, key string) (float64, error) {
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return 0, res.Err()
	}

	return res.Float64()
}

func (s *Service) GetBytes(ctx context.Context, key string) ([]byte, error) {
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return []byte{}, res.Err()
	}

	return res.Bytes()
}

func (s *Service) GetObjectID(ctx context.Context, key string) (primitive.ObjectID, error) {
	res := s.driver.Get(ctx, key)
	if res.Err() != nil {
		return primitive.NilObjectID, res.Err()
	}

	return primitive.ObjectIDFromHex(res.String())
}

func (s *Service) Exists(ctx context.Context, key string) bool {
	res, err := s.driver.Exists(ctx, key).Result()
	if err != nil {
		s.log.From("cache", "exists").Error(err.Error())
		return false
	}
	return res == 1
}

func (s *Service) Keys(ctx context.Context, pattern string) ([]string, error) {
	return s.driver.Keys(ctx, pattern).Result()
}
