package store_test

import (
	"testing"

	"git.martianoids.com/martianoids/martian-stack/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Run("TestNew", func(t *testing.T) {
		s := store.New()
		assert.NotNil(t, s, "Expected a new store instance, got nil")
	})

	t.Run("TestSetAndGet", func(t *testing.T) {
		s := store.New()

		err := s.Set("key1", "value1")
		require.NoError(t, err)

		var result string
		err = s.Get("key1", &result)
		require.NoError(t, err)
		assert.Equal(t, "value1", result, "Expected 'value1'")
	})

	t.Run("TestDelete", func(t *testing.T) {
		s := store.New()

		err := s.Set("key1", "value1")
		require.NoError(t, err)

		s.Delete("key1")

		var result string
		err = s.Get("key1", &result)
		assert.ErrorIs(t, err, store.ErrKeyNotFound, "Expected error due to missing key")
	})

	t.Run("TestIsDirty", func(t *testing.T) {
		s := store.New()
		assert.False(t, s.IsDirty(), "Expected store to be not dirty initially")

		err := s.Set("key1", "value1")
		require.NoError(t, err)
		assert.True(t, s.IsDirty(), "Expected store to be dirty after setting a value")
	})

	t.Run("TestFlush", func(t *testing.T) {
		s := store.New()
		err := s.Set("key1", "value1")
		require.NoError(t, err)

		s.Flush()
		require.NoError(t, err)
		assert.True(t, s.IsDirty(), "Expected store to be clean after flushing")

		var result string
		err = s.Get("key1", &result)
		assert.ErrorIs(t, err, store.ErrKeyNotFound, "Expected error after flush due to missing key")
	})

	t.Run("TestGetString", func(t *testing.T) {
		s := store.New()
		err := s.Set("key1", "value1")
		require.NoError(t, err)

		result := s.GetString("key1")
		assert.Equal(t, "value1", result, "Expected 'value1'")

		s.Delete("key1")
		assert.Equal(t, "", s.GetString("key1"), "Expected empty string after deletion")
	})

	t.Run("TestGetInt", func(t *testing.T) {
		s := store.New()
		err := s.Set("key1", 42)
		require.NoError(t, err)

		result := s.GetInt("key1")
		assert.Equal(t, 42, result, "Expected 42")

		s.Delete("key1")
		assert.Equal(t, 0, s.GetInt("key1"), "Expected 0 after deletion")
	})

	t.Run("TestGetFloat", func(t *testing.T) {
		s := store.New()
		err := s.Set("key1", 3.14)
		require.NoError(t, err)

		result := s.GetFloat("key1")
		assert.Equal(t, 3.14, result, "Expected 3.14")

		s.Delete("key1")
		assert.Equal(t, 0.0, s.GetFloat("key1"), "Expected 0 after deletion")
	})

	// test marshal and unmarshal
	t.Run("TestMarshalJSON", func(t *testing.T) {
		s := store.New()

		require.NoError(t, s.Set("key1", "value1"))
		require.NoError(t, s.Set("key2", 2))

		b, err := s.MarshalJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, b, "Expected non-empty byte slice")
		s.Flush()

		require.NoError(t, s.UnmarshalJSON(b))
		assert.Equal(t, "value1", s.GetString("key1"))
		assert.Equal(t, 2, s.GetInt("key2"))
	})
}
