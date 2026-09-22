package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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
	r.Route("/{id}", func(r chi.Router) {
		r.Use(rs.UserCtx)
		r.Get("/", rs.FindOne)
		r.Put("/", rs.Update)
		r.Delete("/", rs.Delete)
	})

	return r
}

func (rs studentsResource) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var student *db.Student
		var err error

		studentIDStr := chi.URLParam(r, "id")
		if studentIDStr == "" {
			render.Render(w, r, ErrNotFound())
		}

		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
		}

		student, err = store.FindStudentByID(r.Context(), studentID)
		if err != nil {
			render.Render(w, r, ErrNotFound())
			return
		}

		ctx := context.WithValue(r.Context(), "student", student)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

func (rs studentsResource) FindOne(w http.ResponseWriter, r *http.Request) {
	student := r.Context().Value("student").(*db.Student)
	render.JSON(w, r, student)
}

func (rs studentsResource) Update(w http.ResponseWriter, r *http.Request) {
	student := r.Context().Value("student").(*db.Student)

	var req db.Student
	if err := render.Decode(r, &req); err != nil {
		render.Render(w, r, ErrInvalidRequest("invalid json payload"))
		return
	}

	// TODO: Validation

	student = &req
	if err := store.UpdateStudent(r.Context(), student.ID, req); err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	bytes, err := json.Marshal(student)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	hub.Broadcast(websockets.Event{
		Event:   "student:update",
		Payload: bytes,
	})

	render.Status(r, 200)
	render.JSON(w, r, student)
}

func (rs studentsResource) Delete(w http.ResponseWriter, r *http.Request) {
	student := r.Context().Value("student").(*db.Student)

	err := store.DeleteStudentByID(r.Context(), student.ID)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	bytes, err := json.Marshal(student)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	hub.Broadcast(websockets.Event{
		Event:   "student:delete",
		Payload: bytes,
	})

	render.Status(r, 200)
	render.JSON(w, r, student)
}
