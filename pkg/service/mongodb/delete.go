package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) DeleteOne(ctx context.Context, e Entity, id primitive.ObjectID) error {
	col, err := s.Collection(e)
	if err != nil {
		return err
	}

	_, err = col.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}

	exists, err := s.cacheExistsByID(ctx, e, id)
	if err != nil {
		return err
	}
	if exists {
		return s.cacheDelByID(ctx, e, id)
	}

	return nil
}

func (s *Service) DropAll(ctx context.Context, e Entity) error {
	col, err := s.Collection(e)
	if err != nil {
		return err
	}

	keyPattern, err := s.cacheKeyPattern(e)
	if err != nil {
		return err
	}

	// clean cache
	if err := s.cache.DeletePattern(ctx, keyPattern); err != nil {
		return err
	}

	return col.Drop(ctx)
}
