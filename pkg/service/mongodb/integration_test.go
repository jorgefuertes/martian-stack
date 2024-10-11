package mongodb_test

import (
	"context"
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/helper"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/cache/redis"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"
	"git.martianoids.com/martianoids/martian-stack/pkg/service/mongodb"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TestAccount struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username"     bson:"username"      validate:"required,min=3,max=50"`
	Name     string             `json:"name"         bson:"name"          validate:"required,min=3,max=50"`
	Surname  string             `json:"surname"      bson:"surname"       validate:"required,min=3,max=50"`
	Email    string             `json:"email"        bson:"email"         validate:"required,email"`
	Role     string             `json:"role"         bson:"role"          validate:"required,oneof=superadmin admin user" default:"user"`
	Enabled  bool               `json:"enabled"      bson:"enabled"                                                       default:"true"`
}

func NewTestAccount() *TestAccount {
	a := new(TestAccount)
	_ = defaults.Set(a)

	return a
}

func (t *TestAccount) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(t)
}

func (t *TestAccount) Indexes() []mongo.IndexModel {
	return []mongo.IndexModel{{Keys: "email"}}
}

func TestMongoDB(t *testing.T) {
	wr := helper.NewWriter()
	logSvc := logger.New(wr, logger.JsonFormat, logger.LevelDebug)
	cacheSvc := redis.New(logSvc, "localhost", 6379, "", "", 0)
	dbSvc, err := mongodb.NewService(cacheSvc, logSvc, "martian-data-test", "localhost", 27017, 8, 16, "", "")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	testAccount := &TestAccount{
		Username: "john.doe",
		Name:     "John",
		Surname:  "Doe",
		Email:    "john.doe@example.test",
		Role:     "admin",
		Enabled:  true,
	}

	t.Cleanup(func() {
		err := dbSvc.DropAll(ctx, &TestAccount{})
		assert.NoError(t, err)
		err = dbSvc.Close()
		assert.NoError(t, err)
		cancel()
	})

	t.Run("Indexes", func(t *testing.T) {
		err := dbSvc.Indexes(ctx, &TestAccount{})
		require.NoError(t, err)
	})

	t.Run("CreateOne", func(t *testing.T) {
		id, err := dbSvc.CreateOne(ctx, testAccount)
		require.NoError(t, err)
		assert.False(t, id.IsZero())
		t.Run("FindOneByID", func(t *testing.T) {
			found := NewTestAccount()
			err := dbSvc.FindOneByID(ctx, found, id)
			require.NoError(t, err)
			assert.EqualValues(t, testAccount, found)
		})
		t.Run("UpdateOne", func(t *testing.T) {
			testAccount.Name = "Johnny"
			err := dbSvc.UpdateOne(ctx, testAccount)
			require.NoError(t, err)
			t.Run("FindOneByID", func(t *testing.T) {
				found := NewTestAccount()
				err := dbSvc.FindOneByID(ctx, found, id)
				require.NoError(t, err)
				assert.EqualValues(t, testAccount, found)
			})
		})

		t.Run("UpdateOneRaw", func(t *testing.T) {
			testAccount.Name = "John"
			err := dbSvc.UpdateOneRaw(ctx, testAccount, bson.D{{Key: "surname", Value: "Walker"}})
			require.NoError(t, err)
			assert.EqualValues(t, "Johnny", testAccount.Name)
			assert.EqualValues(t, "Walker", testAccount.Surname)
			t.Run("FindOneByID", func(t *testing.T) {
				found := NewTestAccount()
				err := dbSvc.FindOneByID(ctx, found, id)
				require.NoError(t, err)
				assert.EqualValues(t, testAccount, found)
			})
		})

		t.Run("DeleteOne", func(t *testing.T) {
			// create
			newAccount := NewTestAccount()
			newAccount.Name = "John2"
			newAccount.Surname = "Doe2"
			newAccount.Username = "john2.doe2"
			newAccount.Email = "john2@example.com"
			id, err := dbSvc.CreateOne(ctx, newAccount)
			require.NoError(t, err)
			assert.False(t, id.IsZero())

			// find it
			found := NewTestAccount()
			err = dbSvc.FindOneByID(ctx, found, id)
			require.NoError(t, err)
			assert.EqualValues(t, newAccount, found)

			// delete
			err = dbSvc.DeleteOne(ctx, &TestAccount{}, id)
			require.NoError(t, err)

			// try to find it again
			found = NewTestAccount()
			err = dbSvc.FindOneByID(ctx, found, id)
			require.Error(t, err)
			assert.ErrorIs(t, err, mongo.ErrNoDocuments)
		})
	})
}
