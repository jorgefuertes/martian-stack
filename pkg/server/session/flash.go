package session

import "slices"

type FlashLevel string

const (
	flashesKey        string     = "flashes"
	FlashLevelInfo    FlashLevel = "info"
	FlashLevelSuccess FlashLevel = "success"
	FlashLevelWarn    FlashLevel = "warn"
	FlashLevelError   FlashLevel = "error"
)

type Flash struct {
	Level FlashLevel `json:"level"`
	Msg   string     `json:"msg"`
}

func (f Flash) IsEmpty() bool {
	return f.Level == "" && f.Msg == ""
}

type Flashes []Flash

// GetFlashes returns all the flashes in the session, deleting them
func (s *Session) GetAllFlashes() Flashes {
	flashes := Flashes{}
	_ = s.store.Get(flashesKey, &flashes)
	s.store.Delete(flashesKey)

	return flashes
}

// AddFlash adds a flash to the session
func (s *Session) AddFlash(level FlashLevel, msg string) {
	if level == "" {
		level = FlashLevelInfo
	}

	if msg == "" {
		return
	}

	s.store.Set(flashesKey, append(s.GetAllFlashes(), Flash{Level: level, Msg: msg}))
}

func (s *Session) HasFlashes() bool {
	flashes := Flashes{}
	err := s.store.Get(flashesKey, &flashes)
	if err != nil {
		return false
	}

	return len(flashes) > 0
}

func (s *Session) GetNextFlash() Flash {
	flashes := Flashes{}
	_ = s.store.Get(flashesKey, &flashes)
	if len(flashes) == 0 {
		return Flash{}
	}
	f := flashes[0]

	flashes = slices.Delete(flashes, 0, 1)
	_ = s.store.Set(flashesKey, flashes)

	return f
}
