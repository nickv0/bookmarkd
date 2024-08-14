package routes

import (
	"net/http"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

type BookmarksGetResponse struct {
	Bookmarks []*core.Bookmark `json:"bookmarks"`
	N         int              `json:"n"`
}

func handleBookmarksGet(
	bookmarkStore core.BookmarkStore,
) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		filter, err := encoder.DecodeJson[core.BookmarkFilter](r)
		if r.Body != nil && err != nil {
			encoder.EncodeError(w, r, err)
			return
		}

		bookmarks, n, err := bookmarkStore.FindBookmarks(r.Context(), filter)
		if err != nil {
			encoder.EncodeError(w, r, err)
			return
		}

		if err := encoder.EncodeJson(w, http.StatusOK, &BookmarksGetResponse{
			Bookmarks: bookmarks,
			N:         n,
		}); err != nil {
			encoder.EncodeError(w, r, err)
		}
	})
}
