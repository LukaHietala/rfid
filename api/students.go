package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

type studentsResource struct{}

func (rs studentsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)
	r.Post("/", rs.Create)

	return r
}

func (rs studentsResource) List(w http.ResponseWriter, r *http.Request) {
	students, err := store.ListStudents(r.Context())
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}
	render.JSON(w, r, students)
}

func (rs studentsResource) Create(w http.ResponseWriter, r *http.Request) {
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
}
