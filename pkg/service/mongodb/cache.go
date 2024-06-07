package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) cacheKeyPattern(e Entity) (string, error) {
	col, err := getCollectionName(e)
	if err != nil {
		return "", err
	}

	return col + "-*", nil
}

func (s *Service) cacheKeyFromID(e Entity, id primitive.ObjectID) (string, error) {
	keyID := id.Hex()
	if id.IsZero() {
		keyID = primitive.NewObjectID().Hex()
	}

	col, err := getCollectionName(e)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", col, keyID), nil
}

func (s *Service) cacheExists(ctx context.Context, e Entity) (bool, error) {
	id, err := getID(e)
	if err != nil {
		return false, err
	}

	key, err := s.cacheKeyFromID(e, id)
	if err != nil {
		return false, err
	}

	return s.cache.Exists(ctx, key), nil
}

func (s *Service) cacheExistsByID(ctx context.Context, e Entity, id primitive.ObjectID) (bool, error) {
	key, err := s.cacheKeyFromID(e, id)
	if err != nil {
		return false, err
	}

	return s.cache.Exists(ctx, key), nil
}

func (s *Service) cacheSet(ctx context.Context, e Entity) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	id, err := getID(e)
	if err != nil {
		return err
	}

	key, err := s.cacheKeyFromID(e, id)
	if err != nil {
		return err
	}

	return s.cache.Set(ctx, key, b, 24*time.Hour)
}

func (s *Service) cacheGetByID(ctx context.Context, e Entity, id primitive.ObjectID) error {
	key, err := s.cacheKeyFromID(e, id)
	if err != nil {
		return err
	}

	if !s.cache.Exists(ctx, key) {
		return ErrNotFoundInCache
	}
	b, err := s.cache.GetBytes(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, e)
}

func (s *Service) cacheDelByID(ctx context.Context, e Entity, id primitive.ObjectID) error {
	key, err := s.cacheKeyFromID(e, id)
	if err != nil {
		return err
	}

	return s.cache.Delete(ctx, key)
}
