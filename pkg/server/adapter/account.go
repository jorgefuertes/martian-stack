package adapter

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

var (
	ErrPasswordTooShort = fmt.Errorf("password too short, minimum %d characters", minPasswordLength)
	ErrPasswordTooLong  = fmt.Errorf("password too long, maximum %d characters", maxPasswordLength)
)

type Account struct {
	ID              string    `json:"id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastLogin       time.Time `json:"last"`
	Username        string    `json:"username"     validate:"required,min=4,max=50"`
	Name            string    `json:"name"         validate:"required,min=3,max=120"`
	Email           string    `json:"email"        validate:"required,email"`
	Enabled         bool      `json:"enabled"`
	Role            string    `json:"role"         validate:"required,min=3,max=10"  default:"user"`
	CryptedPassword []byte    `json:"_"`
}

func (a Account) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(a)
}

func (a *Account) SetPassword(password string) error {
	bPassword := []byte(password)

	if len(bPassword) < minPasswordLength {
		return ErrPasswordTooShort
	}

	if len(bPassword) > maxPasswordLength {
		return ErrPasswordTooLong
	}

	c, err := bcrypt.GenerateFromPassword(bPassword, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	a.CryptedPassword = c

	return nil
}

func (a *Account) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword(a.CryptedPassword, []byte(password))
}
