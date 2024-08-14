package mock

import "bookmarkd/internal/core"

var _ core.RegistrationStore = (*RegistrationStore)(nil)

type RegistrationStore struct {
	StartRegistrationSessionFn    func(username string) (*core.Registration, error)
	FindRegistrationSessionByIDFn func(id string) (*core.Registration, error)
	DeleteRegistrationSessionFn   func(sessionID string) error
}

func (s *RegistrationStore) StartRegistration(username string) (*core.Registration, error) {
	return s.StartRegistrationSessionFn(username)
}

func (s *RegistrationStore) FindRegistrationByID(id string) (*core.Registration, error) {
	return s.FindRegistrationSessionByIDFn(id)
}

func (s *RegistrationStore) DeleteRegistration(sessionID string) error {
	return s.DeleteRegistrationSessionFn(sessionID)
}
