package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type scanResource struct{}

func (rs scanResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)

	return r
}

func (rs scanResource) List(w http.ResponseWriter, r *http.Request) {
	scans, err := store.ListScans(r.Context())
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}
	render.JSON(w, r, scans)
}
