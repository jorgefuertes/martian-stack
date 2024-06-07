package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Entity interface {
	Validate() error
	Indexes() []mongo.IndexModel
}
