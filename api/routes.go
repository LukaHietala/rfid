package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

var store *db.Store
var hub *websockets.Hub

func NewRouter(s *db.Store, h *websockets.Hub) *chi.Mux {
	// Asiatonta
	store = s
	hub = h

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(hub, w, r)
	})

	r.Route("/api", func(r chi.Router) {
		r.Mount("/students", studentsResource{}.Routes())
		r.Mount("/devices", devicesResource{}.Routes())
		r.Mount("/scans", scanResource{}.Routes())
	})

	return r
}

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	ErrorText      string `json:"error"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(msg string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		ErrorText:      msg,
	}
}

func ErrInternal(err error) render.Renderer {
	log.Println("error:", err)
	return &ErrResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		ErrorText:      http.StatusText(http.StatusInternalServerError),
	}
}

func ErrNotFound() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusNotFound,
		ErrorText:      http.StatusText(http.StatusNotFound),
	}
}
