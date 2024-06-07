package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func (s *Service) checkPings() {
	l := s.log.From(Component, ActionPing).With("database", s.db)
	lastOK := true
	lastClientOK := false
	for {
		if s.closed {
			l.Info("closing ping routine")
			return
		}
		if s.conn == nil {
			lastClientOK = false
			time.Sleep(DbClientWait)
			l.Info("waiting for db client")
			continue
		}
		if !lastClientOK {
			lastClientOK = true
			l.Info("client connected")
		}
		ctx, cancel := context.WithTimeout(context.Background(), DbPingTimeout)
		s.err = s.conn.Ping(ctx, readpref.Primary())
		cancel()
		if s.err != nil && lastOK {
			l.Error("database offline")
		}
		if s.err == nil && !lastOK {
			l.Info("ok", "db", s.db)
		}
		lastOK = s.err == nil
		time.Sleep(DbPingDelay)
	}
}

func (c *Service) GetLastCheck() error {
	return c.err
}
