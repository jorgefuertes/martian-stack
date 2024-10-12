package store

// true if the data has been modified
func (s *Service) IsDirty() bool {
	return s.dirty
}

func (s *Service) SetClean() {
	s.dirty = false
}
