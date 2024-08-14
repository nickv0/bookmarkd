package inmem

import (
	"encoding/base32"
	"fmt"
	"sync"
	"time"

	"bookmarkd"
	"bookmarkd/internal/core"

	"github.com/google/uuid"
)

// Ensure type implements interface.
var _ core.RegistrationStore = (*RegistrationStore)(nil)

// RegistrationStore represents a service for Registrationentication attempts
type RegistrationStore struct {
	mu sync.Mutex

	// Registrationentication sessions
	Registrations map[uuid.UUID]*core.Registration
}

// NewRegistrationStore returns a new instance of RegistrationStore.
func NewRegistrationStore() *RegistrationStore {
	return &RegistrationStore{
		Registrations: make(map[uuid.UUID]*core.Registration),
	}
}

func (s *RegistrationStore) StartRegistration(username string) (*core.Registration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	seed := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(r.String()))

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	a := core.Registration{
		ID:        id,
		Username:  username,
		Seed:      seed,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	s.Registrations[a.ID] = &a

	return &a, nil
}

func (s *RegistrationStore) FindRegistrationByID(id string) (*core.Registration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse string to uuid: %w", bookmarkd.ErrNotFound)
	}

	session, ok := s.Registrations[u]
	if !ok {
		return nil, fmt.Errorf("find registration: %w", bookmarkd.ErrNotFound)
	}

	return session, nil
}

func (s *RegistrationStore) DeleteRegistration(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse string to uuid: %w", bookmarkd.ErrNotFound)
	}

	delete(s.Registrations, u)
	return nil
}
