package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) CreateOne(ctx context.Context, e Entity) (primitive.ObjectID, error) {
	id, err := getID(e)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if err := e.Validate(); err != nil {
		return id, errors.Join(ErrDbValidation, err)
	}

	if !id.IsZero() && id.Hex() != DbAdminHexID {
		return id, ErrDbNotZeroID
	}

	if err := setNewID(e); err != nil {
		return id, err
	}

	id, err = getID(e)
	if err != nil {
		return id, err
	}

	col, err := s.Collection(e)
	if err != nil {
		return id, err
	}

	setCreatedAt(e, time.Now())
	setUpdatedAt(e, time.Now())

	res, err := col.InsertOne(ctx, e)

	return res.InsertedID.(primitive.ObjectID), err
}

func (s *Service) UpdateOne(ctx context.Context, e Entity) error {
	id, err := getID(e)
	if err != nil {
		return err
	}

	if id.IsZero() {
		return ErrDbZeroID
	}

	if err := e.Validate(); err != nil {
		return errors.Join(ErrDbValidation, err)
	}

	col, err := s.Collection(e)
	if err != nil {
		return err
	}

	setUpdatedAt(e, time.Now())

	_, err = col.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, bson.D{{Key: "$set", Value: e}})
	if err != nil {
		return err
	}
	return s.cacheSet(ctx, e)
}

func (s *Service) UpdateOneRaw(ctx context.Context, e Entity, update bson.D) error {
	id, err := getID(e)
	if err != nil {
		return err
	}

	if id.IsZero() {
		return ErrDbZeroID
	}

	setUpdatedAt(e, time.Now())

	updateFilter := bson.D{{Key: "_id", Value: id}}
	updateContents := bson.D{{Key: "$set", Value: update}}

	col, err := s.Collection(e)
	if err != nil {
		s.log.From(Component, "UpdateOneRaw").Error(err.Error())
		return err
	}

	_, err = col.UpdateOne(ctx, updateFilter, updateContents)
	if err != nil {
		s.log.From(Component, "UpdateOneRaw").Error(err.Error())
		return err
	}

	err = col.FindOne(ctx, updateFilter).Decode(e)
	if err != nil {
		s.log.From(Component, "UpdateOneRaw").Error(err.Error())
		return err
	}

	err = s.cacheSet(ctx, e)
	if err != nil {
		s.log.From(Component, "UpdateOneRaw").Error(err.Error())
	}

	return err
}
