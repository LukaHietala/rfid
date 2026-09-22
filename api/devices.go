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

type devicesResource struct{}

func (rs devicesResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)
	r.Post("/", rs.Create)
	r.Route("/{id}", func(r chi.Router) {
		r.Use(rs.DeviceCtx)
		r.Get("/", rs.FindOne)
		r.Put("/", rs.Update)
		r.Delete("/", rs.Delete)
	})

	return r
}

func (rs devicesResource) DeviceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var dev *db.Device
		var err error

		deviceIDStr := chi.URLParam(r, "id")
		if deviceIDStr == "" {
			render.Render(w, r, ErrNotFound())
		}

		deviceID, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
		}

		dev, err = store.FindDeviceByID(r.Context(), deviceID)
		if err != nil {
			render.Render(w, r, ErrNotFound())
			return
		}

		ctx := context.WithValue(r.Context(), "device", dev)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rs devicesResource) List(w http.ResponseWriter, r *http.Request) {
	devices, err := store.ListDevices(r.Context())
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}
	render.JSON(w, r, devices)
}

func (rs devicesResource) Create(w http.ResponseWriter, r *http.Request) {
	var dev db.Device
	if err := render.Decode(r, &dev); err != nil {
		render.Render(w, r, ErrInvalidRequest("invalid json payload"))
		return
	}

	// TODO: validation

	if err := store.AddDevice(r.Context(), &dev); err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	bytes, err := json.Marshal(dev)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	hub.Broadcast(websockets.Event{
		Event:   "device:new",
		Payload: bytes,
	})

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dev)
}

func (rs devicesResource) FindOne(w http.ResponseWriter, r *http.Request) {
	dev := r.Context().Value("device").(*db.Device)
	render.JSON(w, r, dev)
}

func (rs devicesResource) Update(w http.ResponseWriter, r *http.Request) {
	dev := r.Context().Value("device").(*db.Device)

	var req db.Device
	if err := render.Decode(r, &req); err != nil {
		render.Render(w, r, ErrInvalidRequest("invalid json payload"))
		return
	}

	// TODO: Validation

	dev = &req
	if err := store.UpdateDevice(r.Context(), dev.ID, req); err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	bytes, err := json.Marshal(dev)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	hub.Broadcast(websockets.Event{
		Event:   "device:update",
		Payload: bytes,
	})

	render.Status(r, 200)
	render.JSON(w, r, dev)
}

func (rs devicesResource) Delete(w http.ResponseWriter, r *http.Request) {
	dev := r.Context().Value("device").(*db.Device)

	err := store.DeleteDeviceByID(r.Context(), dev.ID)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	bytes, err := json.Marshal(dev)
	if err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	hub.Broadcast(websockets.Event{
		Event:   "device:delete",
		Payload: bytes,
	})

	render.Status(r, 200)
	render.JSON(w, r, dev)
}
