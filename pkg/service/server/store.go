package server

import (
	"encoding/json"
	"errors"
	"sync"
)

var ErrStoreKeyNotFound = errors.New("key not found in store")
var ErrStoreCannotConvertToBytes = errors.New("cannot convert to bytes")

type store struct {
	data map[string]any
	lock *sync.Mutex
}

func newStore() *store {
	return &store{data: make(map[string]any), lock: &sync.Mutex{}}
}

func (s *store) Set(key string, value any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[key] = value
}

func (s *store) SetObject(key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.Set(key, b)

	return nil
}

func (s store) Get(key string) (any, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.data[key]

	return v, ok
}

func (s store) GetObject(key string, dest any) error {
	v, ok := s.Get(key)
	if !ok || v == nil {
		return ErrStoreKeyNotFound
	}

	b, ok := v.([]byte)
	if !ok {
		return ErrStoreCannotConvertToBytes
	}

	return json.Unmarshal(b, dest)
}

func (s store) GetString(key string) string {
	v, ok := s.Get(key)
	if !ok || v == nil {
		return ""
	}

	return v.(string)
}

func (s store) GetInt(key string) int {
	v, ok := s.Get(key)
	if !ok || v == nil {
		return 0
	}

	return v.(int)
}
