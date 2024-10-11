package store

import (
	"encoding/json"
	"sync"
)

type Service struct {
	data  map[string][]byte
	dirty bool
	lock  *sync.Mutex
}

func New() *Service {
	s := &Service{
		data:  make(map[string][]byte),
		dirty: false,
		lock:  &sync.Mutex{},
	}

	return s
}

// flushes all the data
func (s *Service) Flush() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = make(map[string][]byte)
	s.dirty = true
}

// true if the data has been modified
func (s *Service) IsDirty() bool {
	return s.dirty
}

// set any value
func (s *Service) Set(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = b
	s.dirty = true

	return nil
}

// get any value into dest
func (s *Service) Get(key string, dest any) error {
	v, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}

	return json.Unmarshal(v, dest)
}

// delete a key
func (s *Service) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
	s.dirty = true
}

// get as string, empty if not found
func (s *Service) GetString(key string) string {
	var v string
	_ = s.Get(key, &v)
	return v
}

// get as int, 0 if not found
func (s *Service) GetInt(key string) int {
	var v int
	_ = s.Get(key, &v)
	return v
}

// get as float, 0 if not found
func (s *Service) GetFloat(key string) float64 {
	var v float64
	_ = s.Get(key, &v)
	return v
}
