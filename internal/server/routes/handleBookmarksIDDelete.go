package routes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

func handleBookmarksIDDelete(
	bookmarkStore core.BookmarkStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			p := chi.URLParam(r, "id")

			id, err := strconv.Atoi(p)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrNotFound)
			}

			b, err := bookmarkStore.DeleteBookmark(r.Context(), id)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if err := encoder.EncodeJson(w, http.StatusOK, &b); err != nil {
				encoder.EncodeError(w, r, err)
			}
		})
}
