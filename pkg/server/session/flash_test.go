package session_test

import (
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/server/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlash(t *testing.T) {
	s := session.New()

	testFlashes := session.Flashes{
		{Level: session.FlashLevelInfo, Msg: "Info test"},
		{Level: session.FlashLevelSuccess, Msg: "Success test"},
		{Level: session.FlashLevelWarn, Msg: "Warn test"},
		{Level: session.FlashLevelError, Msg: "Error test"},
	}

	setFlashes := func() {
		for _, f := range testFlashes {
			s.AddFlash(f.Level, f.Msg)
		}
	}

	t.Run("GetAllFlashes", func(t *testing.T) {
		setFlashes()
		require.True(t, s.HasFlashes())
		flashes := s.GetAllFlashes()
		assert.Equal(t, testFlashes, flashes)
		assert.False(t, s.HasFlashes())
	})

	t.Run("GetNextFlash", func(t *testing.T) {
		setFlashes()
		require.True(t, s.HasFlashes())
		count := 0
		for {
			flash := s.GetNextFlash()
			if flash.IsEmpty() {
				break
			}
			count++
		}
		assert.Equal(t, len(testFlashes), count)
		assert.False(t, s.HasFlashes())
	})
}
