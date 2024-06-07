package mongodb

import (
	"context"
	"fmt"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/cache"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	db     string
	cache  cache.Service
	log    *logger.Service
	conn   *mongo.Client
	err    error
	closed bool
}

func NewService(
	cacheSvc cache.Service,
	logSvc *logger.Service,
	database string,
	host string, port int,
	minPoolSize int,
	maxPoolSize int,
	user, pass string,
) (*Service, error) {
	sublog := logSvc.From("dbservice", "new").With("db", database)
	c := &Service{db: database, cache: cacheSvc, closed: false, log: logSvc}
	go c.checkPings()

	o := new(options.ClientOptions).
		SetHosts([]string{fmt.Sprintf("%s:%d", host, port)}).
		SetAppName(database).
		SetConnectTimeout(DbConnTimeout).
		SetTimeout(DbOperationTimeout).
		SetMinPoolSize(uint64(minPoolSize)).
		SetMaxPoolSize(uint64(maxPoolSize))

	if user != "" {
		o.SetAuth(options.Credential{Username: user, Password: pass})
	}

	ctx, cancel := context.WithTimeout(context.Background(), DbOperationTimeout)
	defer cancel()

	c.conn, c.err = mongo.Connect(ctx, o)
	if c.err != nil {
		sublog.Error(ErrDbConn.Error())
	} else {
		sublog.Info("connected to database")
	}
	return c, c.err
}

func (c *Service) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), DbConnTimeout)
	defer cancel()
	c.closed = true
	if err := c.conn.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Service) Conn() (*mongo.Client, error) {
	if c.conn == nil {
		return nil, ErrDbClientNotFound
	}

	return c.conn, nil
}

func (c *Service) Collection(e Entity) (*mongo.Collection, error) {
	if c.conn == nil {
		return nil, ErrDbClientNotFound
	}

	col, err := getCollectionName(e)
	if err != nil {
		return nil, err
	}

	return c.conn.Database(c.db).Collection(col), nil
}

func (c *Service) ConvertID(id any) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case string:
		return primitive.ObjectIDFromHex(v)
	case primitive.ObjectID:
		return v, nil
	}
	return primitive.NilObjectID, ErrDbClientNotFound
}
