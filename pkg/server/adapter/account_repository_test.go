package adapter_test

import (
	"testing"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"github.com/stretchr/testify/require"
)

func TestAccountRepository(t *testing.T) {
	r := adapter.NewInMemoryAccountRepository()

	accounts := []*adapter.Account{
		{Username: "test", Name: "Test", Email: "test@test.com", Role: "user", Enabled: true},
		{Username: "test2", Name: "Test Two", Email: "test2@test.com", Role: "user", Enabled: false},
		{Username: "test3", Name: "Test Three", Email: "test3@test.com", Role: "user", Enabled: true},
	}

	// create
	for _, a := range accounts {
		a.SetPassword(a.Username + "-password")
		err := r.Create(a)
		require.NoError(t, err)
		require.NotEmpty(t, a.ID)
	}

	// get by id
	for _, a := range accounts {
		a2, err := r.Get(a.ID)
		require.NoError(t, err)
		require.Equal(t, a, a2)
	}

	// get by email
	for _, a := range accounts {
		a2, err := r.GetByEmail(a.Email)
		require.NoError(t, err)
		require.Equal(t, a, a2)
	}

	// get by username
	for _, a := range accounts {
		a2, err := r.GetByUsername(a.Username)
		require.NoError(t, err)
		require.Equal(t, a, a2)
	}

	// update
	for _, a := range accounts {
		a.Name = "Updated " + a.Name
		err := r.Update(a)
		require.NoError(t, err)
	}
	for _, a := range accounts {
		a2, err := r.Get(a.ID)
		require.NoError(t, err)
		require.Equal(t, a, a2)
	}

	// delete
	for _, a := range accounts {
		err := r.Delete(a.ID)
		require.NoError(t, err)
	}
	for _, a := range accounts {
		a2, err := r.Get(a.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, adapter.ErrAccountNotFound)
		require.Nil(t, a2)
	}

	// create with ID
	accounts[0].ID = "test-id"
	err := r.Create(accounts[0])
	require.Error(t, err)
	require.ErrorIs(t, err, adapter.ErrCannotCreateWithID)
}
