package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
		render.Render(w, r, ErrInvalidRequest("Invalid json payload"))
		return
	}

	if s.Status == "" {
		s.Status = "OUT"
	}

	decoded, err := hex.DecodeString(s.UID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest("Invalid uid format, unable to decode string to bytes"))
		return
	}

	if len(decoded) != 4 && len(decoded) != 7 {
		render.Render(w, r, ErrInvalidRequest("RFID uid length must be either 4 or 7"))
		return
	}

	if len(s.Name) < 1 || len(s.Name) > 255 {
		render.Render(w, r, ErrInvalidRequest("Name length must be between 1-255"))
		return
	}

	start, err := time.Parse(time.DateOnly, s.StartDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest("Start date is in invalid format"))
		return
	}

	end, err := time.Parse(time.DateOnly, s.EndDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest("End date is in invalid format"))
		return
	}

	if !end.After(start) {
		render.Render(w, r, ErrInvalidRequest("End date must be after start date"))
		return
	}

	for weekday, day := range s.Schedule {
		if day == nil {
			continue
		}
		dayStarts, err := time.Parse(time.TimeOnly, day.Start)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(fmt.Sprintf("Start time is in invalid format at weekday %d", weekday)))
			return
		}
		dayEnds, err := time.Parse(time.TimeOnly, day.End)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(fmt.Sprintf("End time is in invalid format at weekday %d", weekday)))
			return
		}

		if !dayEnds.After(dayStarts) {
			render.Render(w, r, ErrInvalidRequest("End time must be after start time"))
			return
		}
	}

	for _, excluded := range s.ExcludedDays {
		if _, err := time.Parse(time.DateOnly, excluded); err != nil {
			render.Render(w, r, ErrInvalidRequest("Excluded day is in invalid format"))
			return
		}
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
		render.Render(w, r, ErrInvalidRequest("Invalid json payload"))
		return
	}

	decoded, err := hex.DecodeString(req.UID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest("Invalid uid format, unable to decode string to bytes"))
		return
	}

	if len(decoded) != 4 && len(decoded) != 7 {
		render.Render(w, r, ErrInvalidRequest("RFID uid length must be either 4 or 7"))
		return
	}

	if len(req.Name) < 1 || len(req.Name) > 255 {
		render.Render(w, r, ErrInvalidRequest("Name length must be between 1-255"))
		return
	}

	start, err := time.Parse(time.DateOnly, req.StartDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest("Start date is in invalid format"))
		return
	}

	end, err := time.Parse(time.DateOnly, req.EndDate)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest("End date is in invalid format"))
		return
	}

	if !end.After(start) {
		render.Render(w, r, ErrInvalidRequest("End date must be after start date"))
		return
	}

	for weekday, day := range req.Schedule {
		if day == nil {
			continue
		}
		dayStarts, err := time.Parse(time.TimeOnly, day.Start)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(fmt.Sprintf("Start time is in invalid format at weekday %d", weekday)))
			return
		}
		dayEnds, err := time.Parse(time.TimeOnly, day.End)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(fmt.Sprintf("End time is in invalid format at weekday %d", weekday)))
			return
		}

		if !dayEnds.After(dayStarts) {
			render.Render(w, r, ErrInvalidRequest("End time must be after start time"))
			return
		}
	}

	for _, excluded := range req.ExcludedDays {
		if _, err := time.Parse(time.DateOnly, excluded); err != nil {
			render.Render(w, r, ErrInvalidRequest("Excluded day is in invalid format"))
			return
		}
	}

	student = &req
	if err := store.UpdateStudent(r.Context(), student.ID, &req); err != nil {
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
