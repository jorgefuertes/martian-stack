package adapter

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Account interface {
	GetID() primitive.ObjectID
	GetUsername() string
	GetName() string
	GetEmail() string
	GetRole() string
	IsEnabled() bool
}

type MinAccount struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username"     bson:"username"      validate:"required,min=3,max=50"`
	Name     string             `json:"name"         bson:"name"          validate:"required,min=3,max=50"`
	Surname  string             `json:"surname"      bson:"surname"       validate:"required,min=3,max=50"`
	Email    string             `json:"email"        bson:"email"         validate:"required,email"`
	Role     string             `json:"role"         bson:"role"          validate:"required,oneof=superadmin admin user" default:"user"`
	Enabled  bool               `json:"enabled"      bson:"enabled"                                                       default:"true"`
}

type Alias MinAccount

func (a MinAccount) GetUsername() string {
	return a.Username
}

func (a MinAccount) GetName() string {
	return a.Name
}

func (a MinAccount) GetEmail() string {
	return a.Email
}

func (a MinAccount) GetRole() string {
	return a.Role
}

func (a MinAccount) IsEnabled() bool {
	return a.Enabled
}

func (a MinAccount) DbCollName() string {
	return "accounts"
}

func (a MinAccount) GetID() primitive.ObjectID {
	return a.ID
}

func (a *MinAccount) NewID() {}

func (a *MinAccount) SetCreatedAt() {}

func (a *MinAccount) SetUpdatedAt() {}

func (a *MinAccount) Indexes() []mongo.IndexModel {
	return make([]mongo.IndexModel, 0)
}

func (a *MinAccount) UnmarshalBSON(data []byte) error {
	if err := bson.Unmarshal(data, (*Alias)(a)); err != nil {
		return err
	}

	return nil
}

func (a *MinAccount) MarshalBSON() ([]byte, error) {
	var err error

	encoded, err := bson.Marshal((*Alias)(a))
	return encoded, err
}

func (a *MinAccount) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*Alias)(a)); err != nil {
		return err
	}

	return nil
}

func (a *MinAccount) MarshalJSON() ([]byte, error) {
	var err error

	encoded, err := json.Marshal((*Alias)(a))
	return encoded, err
}
