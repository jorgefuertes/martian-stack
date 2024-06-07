package mongodb

import (
	"errors"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/gobeam/stringy"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getCollectionName(e any) (string, error) {
	if !structs.IsStruct(e) {
		return "", ErrReflectNotStruct
	}

	return stringy.New(structs.Name(e)).PascalCase().Get(), nil
}

func getFieldByTag(e any, tag string) (*structs.Field, error) {
	if !structs.IsStruct(e) {
		return nil, ErrReflectNotStruct
	}

	for _, f := range structs.Fields(e) {
		if !f.IsExported() {
			continue
		}
		if strings.HasPrefix(f.Tag("bson"), tag) {
			return f, nil
		}
	}

	return nil, ErrTagNotFound
}

func getID(e any) (primitive.ObjectID, error) {
	f, err := getFieldByTag(e, "_id")
	if err != nil {
		if errors.Is(err, ErrTagNotFound) {
			return primitive.NilObjectID, ErrMissingIDField
		}

		return primitive.NilObjectID, err
	}

	return f.Value().(primitive.ObjectID), nil
}

func setFieldByTag(e any, tag string, value any) error {
	f, err := getFieldByTag(e, tag)
	if err != nil {
		return err
	}
	return f.Set(value)
}

func setID(e any, id primitive.ObjectID) error {
	err := setFieldByTag(e, "_id", id)
	if errors.Is(err, ErrTagNotFound) {
		return ErrMissingIDField
	}

	return err
}

func setNewID(e any) error {
	return setID(e, primitive.NewObjectID())
}

func setCreatedAt(e any, t time.Time) error {
	err := setFieldByTag(e, "created_at", t)
	if errors.Is(err, ErrTagNotFound) {
		return ErrMissingCreatedAtField
	}

	return err
}

func setUpdatedAt(e any, t time.Time) error {
	err := setFieldByTag(e, "updated_at", t)
	if errors.Is(err, ErrTagNotFound) {
		return ErrMissingUpdatedAtField
	}

	return err
}
