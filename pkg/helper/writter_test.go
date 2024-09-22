package helper_test

import (
	"fmt"
	"io"
	"testing"

	"git.martianoids.com/martianoids/martian-stack/pkg/helper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelperWriter(t *testing.T) {
	testLine := "test hello world"
	w := helper.NewWriter()
	w.Write([]byte(testLine))
	assert.Equal(t, 1, w.Len())
	line, err := w.ReadString()
	require.NoError(t, err)
	assert.Equal(t, testLine, line)
	assert.Zero(t, w.Len())

	nLines := 100
	for i := 1; i < nLines; i++ {
		w.Write([]byte(fmt.Sprintf("%s %d", testLine, i)))
		require.Equal(t, i, w.Len())
	}

	for i := nLines - 1; i > 0; i-- {
		line, err := w.ReadString()
		require.NoError(t, err, "Remaining lines: %d", i)
		assert.Contains(t, line, testLine)
	}

	line, err = w.ReadString()
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
	assert.Empty(t, line)
}

func TestReadJSON(t *testing.T) {
	type User struct {
		Name string
		Age  uint8
		City string
	}

	user := User{Name: "John", Age: 30, City: "New York"}

	w := helper.NewWriter()
	n, err := w.WriteJSON(user)
	require.NoError(t, err)
	require.NotZero(t, n)

	var u User
	err = w.ReadJSON(&u)
	require.NoError(t, err)
	assert.Equal(t, user, u)
}
