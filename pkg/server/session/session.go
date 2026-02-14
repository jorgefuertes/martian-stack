package session

import (
	"github.com/google/uuid"
	"github.com/jorgefuertes/martian-stack/pkg/store"
)

type Session struct {
	ID    string
	store *store.Service
}

func New() *Session {
	return &Session{
		ID:    uuid.NewString(),
		store: store.New(),
	}
}

func (s *Session) WithID(id string) *Session {
	s.ID = id
	return s
}

func (s Session) KeyID() string {
	return "sess:" + s.ID
}

func (s Session) Data() *store.Service {
	return s.store
}

func (s Session) MarshalJSON() ([]byte, error) {
	return s.store.MarshalJSON()
}

func (s Session) UnmarshalJSON(b []byte) error {
	return s.store.UnmarshalJSON(b)
}
