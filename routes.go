package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

func NewRouter(store *db.Store, hub *websockets.Hub) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(hub, w, r)
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/students", func(w http.ResponseWriter, r *http.Request) {
			students, err := store.ListStudents(r.Context())
			if err != nil {
				render.Render(w, r, ErrInternal(err))
				return
			}
			render.JSON(w, r, students)
		})

		r.Post("/students", func(w http.ResponseWriter, r *http.Request) {
			var s db.Student
			if err := render.Decode(r, &s); err != nil {
				render.Render(w, r, ErrInvalidRequest("invalid json payload"))
				return
			}

			if s.Status == "" {
				s.Status = "OUT"
			}

			if err := store.AddStudent(r.Context(), &s); err != nil {
				render.Render(w, r, ErrInternal(err))
				return
			}

			bytes, err := json.Marshal(s)
			if err != nil {
				render.Render(w, r, ErrInternal(err))
				return
			}

			hub.Broadcast(websockets.Event{
				Event:   "student:new",
				Payload: bytes,
			})

			render.Status(r, http.StatusCreated)
			render.JSON(w, r, s)
		})
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
