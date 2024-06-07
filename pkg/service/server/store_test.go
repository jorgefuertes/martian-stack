package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	s := newStore()

	s.Set("test-1", "Test One")
	assert.Equal(t, "Test One", s.GetString("test-1"))

	s.Set("test-2", 2)
	assert.Equal(t, 2, s.GetInt("test-2"))

	s.Set("test-3", 2.1)
	assert.Equal(t, 2, s.GetInt("test-2"))

	type TestObject struct {
		One int
		Hello string
	}

	s.SetObject("test-4", &TestObject{One: 1, Hello: "Hello, world!"})
	dest := new(TestObject)
	require.NoError(t, s.GetObject("test-4", dest))
	assert.Equal(t, dest.One, 1)
	assert.Equal(t, dest.Hello, "Hello, world!")

	bomb := map[string]interface{}{"foo": make(chan int)}
	err := s.SetObject("test-5", bomb)
	require.Error(t, err)

	assert.Empty(t, s.GetString("nonexistent"))
	assert.Zero(t, s.GetInt("nonexistent"))
}
