package mock

import (
	"context"

	"bookmarkd/internal/core"
)

var _ core.UserStore = (*UserStore)(nil)

type UserStore struct {
	FindUserByIDFn       func(ctx context.Context, id string) (*core.User, error)
	FindUserByUsernameFn func(ctx context.Context, username string) (*core.User, error)
	FindUsersFn          func(ctx context.Context, filter core.UserFilter) ([]*core.User, int, error)
	CreateUserFn         func(ctx context.Context, user *core.User) error
	UpdateUserFn         func(ctx context.Context, id string, upd core.UserUpdate) (*core.User, error)
	DeleteUserFn         func(ctx context.Context, id string) (*core.User, error)
}

func (s *UserStore) FindUserByID(ctx context.Context, id string) (*core.User, error) {
	return s.FindUserByIDFn(ctx, id)
}

func (s *UserStore) FindUserByUsername(ctx context.Context, username string) (*core.User, error) {
	return s.FindUserByUsernameFn(ctx, username)
}

func (s *UserStore) FindUsers(ctx context.Context, filter core.UserFilter) ([]*core.User, int, error) {
	return s.FindUsersFn(ctx, filter)
}

func (s *UserStore) CreateUser(ctx context.Context, user *core.User) error {
	return s.CreateUserFn(ctx, user)
}

func (s *UserStore) UpdateUser(ctx context.Context, id string, upd core.UserUpdate) (*core.User, error) {
	return s.UpdateUserFn(ctx, id, upd)
}

func (s *UserStore) DeleteUser(ctx context.Context, id string) (*core.User, error) {
	return s.DeleteUserFn(ctx, id)
}
