package store

// marshal and unmarshal json

import (
	"encoding/json"
)

func (s *Service) MarshalJSON() ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	b, err := json.Marshal(s.data)
	if err != nil {
		return nil, err
	}

	s.dirty = false

	return b, nil
}

func (s *Service) UnmarshalJSON(b []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = make(map[string][]byte)
	s.dirty = false

	return json.Unmarshal(b, &s.data)
}
