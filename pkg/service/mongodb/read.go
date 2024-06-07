package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Service) FindOne(ctx context.Context, e Entity, filter bson.D) error {
	col, err := c.Collection(e)
	if err != nil {
		return err
	}

	err = col.FindOne(ctx, filter).Decode(e)
	if err != nil {
		return err
	}

	exists, err := c.cacheExists(ctx, e)
	if err != nil {
		return err
	}

	if !exists {
		return c.cacheSet(ctx, e)
	}

	return nil
}

func (c *Service) FindOneByID(ctx context.Context, e Entity, id primitive.ObjectID) error {
	// decode from redis if found
	if c.cacheGetByID(ctx, e, id) == nil {
		return nil
	}

	// find in DB
	err := c.FindOne(ctx, e, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}

	// store it in redis for 1 hour
	return c.cacheSet(ctx, e)
}

func (c *Service) Find(ctx context.Context, e Entity, filter bson.D) (*mongo.Cursor, error) {
	col, err := c.Collection(e)
	if err != nil {
		return nil, err
	}

	return col.Find(ctx, filter)
}

func (c *Service) FindOneWithOptions(ctx context.Context, e Entity, filter bson.D, o *options.FindOneOptions) error {
	col, err := c.Collection(e)
	if err != nil {
		return err
	}
	return col.FindOne(ctx, filter, o).Decode(e)
}

func (c *Service) FindWithOptions(
	ctx context.Context,
	e Entity,
	filter bson.D,
	o *options.FindOptions,
) (*mongo.Cursor, error) {
	col, err := c.Collection(e)
	if err != nil {
		return nil, err
	}

	return col.Find(ctx, filter, o)
}

func (c *Service) Count(ctx context.Context, e Entity, filter bson.D) (int64, error) {
	col, err := c.Collection(e)
	if err != nil {
		return 0, err
	}

	return col.CountDocuments(ctx, filter)
}

func (c *Service) Exists(ctx context.Context, e Entity, filter bson.D) (bool, error) {
	n, err := c.Count(ctx, e, filter)
	return n > 0, err
}
