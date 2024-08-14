package mock

import (
	"context"

	"bookmarkd/internal/core"
)

var _ core.BookmarkStore = (*BookmarkStore)(nil)

type BookmarkStore struct {
	FindBookmarkByIDFn func(ctx context.Context, id int) (*core.Bookmark, error)
	FindBookmarksFn    func(ctx context.Context, filter core.BookmarkFilter) ([]*core.Bookmark, int, error)
	CreateBookmarkFn   func(ctx context.Context, bookmark *core.Bookmark) error
	UpdateBookmarkFn   func(ctx context.Context, id int, update core.BookmarkUpdate) (*core.Bookmark, error)
	DeleteBookmarkFn   func(ctx context.Context, id int) (*core.Bookmark, error)
}

func (s *BookmarkStore) FindBookmarkByID(ctx context.Context, id int) (*core.Bookmark, error) {
	return s.FindBookmarkByIDFn(ctx, id)
}

func (s *BookmarkStore) FindBookmarks(ctx context.Context, filter core.BookmarkFilter) ([]*core.Bookmark, int, error) {
	return s.FindBookmarksFn(ctx, filter)
}

func (s *BookmarkStore) CreateBookmark(ctx context.Context, bookmark *core.Bookmark) error {
	return s.CreateBookmarkFn(ctx, bookmark)
}

func (s *BookmarkStore) UpdateBookmark(ctx context.Context, id int, update core.BookmarkUpdate) (*core.Bookmark, error) {
	return s.UpdateBookmarkFn(ctx, id, update)
}

func (s *BookmarkStore) DeleteBookmark(ctx context.Context, id int) (*core.Bookmark, error) {
	return s.DeleteBookmarkFn(ctx, id)
}
