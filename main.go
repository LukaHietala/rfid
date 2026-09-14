package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/websockets"
)

func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	store := db.NewStore(conn)

	hub := websockets.NewHub()
	go hub.Run()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(hub, w, r)
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/students", func(w http.ResponseWriter, r *http.Request) {
			students, err := store.ListStudents(r.Context())
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(500), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(students)
		})

		r.Post("/students", func(w http.ResponseWriter, r *http.Request) {
			var s db.Student
			if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}

			if s.Status == "" {
				s.Status = "OUT"
			}

			if err := store.AddStudent(r.Context(), &s); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}

			bytes, err := json.Marshal(s)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}

			hub.Broadcast(websockets.Event{
				Event:   "student:new",
				Payload: bytes,
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(bytes)
		})
	})

	log.Println("Server running on :3000")
	http.ListenAndServe(":3000", r)
}
