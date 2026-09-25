package api

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

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

	templateFs := os.DirFS("web/templates/")
	staticFs := os.DirFS("web/static/")
	FileServer(r, "/static", staticFs)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, templateFs, "index.html")
	})

	r.Route("/api", func(r chi.Router) {
		r.Mount("/students", studentsResource{}.Routes())
		r.Mount("/devices", devicesResource{}.Routes())
		r.Mount("/scans", scanResource{}.Routes())
	})

	return r
}

func FileServer(r chi.Router, path string, rootFS fs.FS) {
	root := http.FS(rootFS)
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	ErrorText      string `json:"error"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		ErrorText:      err.Error(),
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
