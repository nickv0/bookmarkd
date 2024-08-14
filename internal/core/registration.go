package core

import (
	"time"

	"github.com/google/uuid"
)

type Registration struct {
	ID        uuid.UUID
	Seed      string
	Username  string
	ExpiresAt time.Time
}

type RegistrationStore interface {
	StartRegistration(username string) (*Registration, error)
	FindRegistrationByID(id string) (*Registration, error)
	DeleteRegistration(id string) error
}

// NopRegistrationStore returns an Registration service that does nothing.
func NopRegistrationStore() RegistrationStore { return &nopRegistrationStore{} }

type nopRegistrationStore struct{}

func (*nopRegistrationStore) StartRegistration(username string) (*Registration, error) {
	panic("not implemented")
}

func (*nopRegistrationStore) FindRegistrationByID(id string) (*Registration, error) {
	panic("not implemented")
}

func (*nopRegistrationStore) DeleteRegistration(id string) error {
	panic("not implemented")
}
