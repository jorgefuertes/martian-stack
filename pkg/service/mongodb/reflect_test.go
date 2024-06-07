package mongodb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEntityReflection(t *testing.T) {
	type TestEntity struct {
		ID        primitive.ObjectID `json:"id"         bson:"_id"`
		Name      string             `json:"name"       bson:"name"`
		pvt       string
		CreatedAt time.Time `json:"created_at" bson:"created_at"`
		UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	}

	type NoIDEntity struct {
		Name string `json:"name" bson:"name"`
	}

	t.Run("getCollectionName", func(t *testing.T) {
		name, err := getCollectionName(TestEntity{})
		require.NoError(t, err)
		assert.Equal(t, "TestEntity", name)

		name, err = getCollectionName(&TestEntity{})
		require.NoError(t, err)
		assert.Equal(t, "TestEntity", name)

		_, err = getCollectionName(map[string]string{})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrReflectNotStruct)

		name, err = getCollectionName(&NoIDEntity{})
		require.NoError(t, err)
		assert.Equal(t, "NoIDEntity", name)
	})

	t.Run("getFieldByTag", func(t *testing.T) {
		e := TestEntity{ID: primitive.NewObjectID(), Name: "John Doe", pvt: ""}
		f, err := getFieldByTag(e, "name")
		require.NoError(t, err)
		assert.Equal(t, f.Value(), e.Name)

		_, err = getFieldByTag(e, "fake_field")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTagNotFound)

		_, err = getFieldByTag(map[string]string{}, "fake_field")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrReflectNotStruct)
	})

	t.Run("getID", func(t *testing.T) {
		e := TestEntity{ID: primitive.NewObjectID(), Name: "John Doe"}
		id, err := getID(e)
		require.NoError(t, err)
		assert.Equal(t, e.ID, id)

		_, err = getID(NoIDEntity{})
		require.Error(t, err, ErrMissingIDField)

		_, err = getID(map[string]string{})
		require.Error(t, err, ErrReflectNotStruct)
	})

	t.Run("setFieldByTag", func(t *testing.T) {
		e := &TestEntity{}
		err := setFieldByTag(e, "name", "John Doe")
		require.NoError(t, err)
		assert.Equal(t, "John Doe", e.Name)

		err = setFieldByTag(e, "fake_field", "John Doe")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTagNotFound)
	})

	t.Run("setID", func(t *testing.T) {
		e := &TestEntity{}
		id := primitive.NewObjectID()
		err := setID(e, id)
		require.NoError(t, err)
		assert.Equal(t, e.ID, id)

		err = setID(NoIDEntity{}, id)
		require.Error(t, err, ErrMissingIDField)

		err = setID(map[string]string{}, id)
		require.Error(t, err, ErrReflectNotStruct)
	})

	t.Run("setNewID", func(t *testing.T) {
		e := &TestEntity{}
		err := setNewID(e)
		require.NoError(t, err)
		assert.NotEqual(t, e.ID, primitive.NilObjectID)
		assert.False(t, e.ID.IsZero())

		err = setNewID(NoIDEntity{})
		require.Error(t, err, ErrMissingIDField)

		err = setNewID(map[string]string{})
		require.Error(t, err, ErrReflectNotStruct)
	})

	t.Run("setCreatedAt", func(t *testing.T) {
		now := time.Now()
		e := &TestEntity{}
		err := setCreatedAt(e, now)
		require.NoError(t, err)
		assert.Equal(t, now, e.CreatedAt)

		err = setCreatedAt(&NoIDEntity{}, now)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingCreatedAtField)
	})

	t.Run("setUpdatedAt", func(t *testing.T) {
		now := time.Now()
		e := &TestEntity{}
		err := setUpdatedAt(e, now)
		require.NoError(t, err)
		assert.Equal(t, now, e.UpdatedAt)

		err = setUpdatedAt(&NoIDEntity{}, now)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingUpdatedAtField)
	})
}
