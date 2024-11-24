package adapter

import (
	"errors"
	"slices"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrCannotCreateWithID = errors.New("cannot create account with ID")
	ErrPasswordNotSet     = errors.New("password not set")
)

type InMemoryAccountRepository struct {
	accounts []Account
	lock     *sync.Mutex
}

func NewInMemoryAccountRepository() *InMemoryAccountRepository {
	return &InMemoryAccountRepository{
		accounts: make([]Account, 0),
		lock:     &sync.Mutex{},
	}
}

func (r *InMemoryAccountRepository) Get(id string) (*Account, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, a := range r.accounts {
		if a.ID == id {
			copyAccount := a
			return &copyAccount, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (r *InMemoryAccountRepository) Exists(id string) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, a := range r.accounts {
		if a.ID == id {
			return true
		}
	}
	return false
}

func (r *InMemoryAccountRepository) GetByEmail(email string) (*Account, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, a := range r.accounts {
		if a.Email == email {
			copyAccount := a
			return &copyAccount, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (r *InMemoryAccountRepository) GetByUsername(email string) (*Account, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, a := range r.accounts {
		if a.Username == email {
			copyAccount := a
			return &copyAccount, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (r *InMemoryAccountRepository) Create(a *Account) error {
	if err := a.Validate(); err != nil {
		return err
	}

	if a.ID != "" {
		return ErrCannotCreateWithID
	}

	if len(a.CryptedPassword) == 0 {
		return ErrPasswordNotSet
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	a.ID = uuid.NewString()
	r.accounts = append(r.accounts, *a)

	return nil
}

func (r *InMemoryAccountRepository) Update(a *Account) error {
	if err := a.Validate(); err != nil {
		return err
	}

	if len(a.CryptedPassword) == 0 {
		return ErrPasswordNotSet
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for i, acc := range r.accounts {
		if acc.ID == a.ID {
			r.accounts[i] = *a
			return nil
		}
	}

	return ErrAccountNotFound
}

func (r *InMemoryAccountRepository) Delete(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for i, acc := range r.accounts {
		if acc.ID == id {
			r.accounts = slices.Delete(r.accounts, i, 1)
			return nil
		}
	}

	return ErrAccountNotFound
}
