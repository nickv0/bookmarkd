package routes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

func handleBookmarksIDPatch(
	bookmarkStore core.BookmarkStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			p := chi.URLParam(r, "id")

			id, err := strconv.Atoi(p)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrNotFound)
			}

			upd, err := encoder.DecodeJson[core.BookmarkUpdate](r)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			b, err := bookmarkStore.UpdateBookmark(r.Context(), id, upd)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if err := encoder.EncodeJson(w, http.StatusOK, &b); err != nil {
				encoder.EncodeError(w, r, err)
			}
		})
}
