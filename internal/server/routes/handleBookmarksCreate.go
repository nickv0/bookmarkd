package routes

import (
	"net/http"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

func handleBookmarksCreate(
	bookmarkStore core.BookmarkStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			b, err := encoder.DecodeJson[core.Bookmark](r)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if err := bookmarkStore.CreateBookmark(r.Context(), &b); err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if err := encoder.EncodeJson(w, http.StatusOK, &b); err != nil {
				encoder.EncodeError(w, r, err)
			}
		})
}
