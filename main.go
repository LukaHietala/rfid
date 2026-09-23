package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lukahietala/rfid/api"
	"github.com/lukahietala/rfid/db"
	"github.com/lukahietala/rfid/tcp"
	"github.com/lukahietala/rfid/websockets"
	"github.com/lukahietala/rfid/workers"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer conn.Close()
	store := db.NewStore(conn)

	hub := websockets.NewHub()
	go hub.Run()

	go workers.StartDoneTicker(ctx, store, hub, 10*time.Second)

	rfidSrv := tcp.New(":5000", store, hub)
	go func() {
		if err := rfidSrv.Start(ctx); err != nil {
			log.Printf("failed to start rfid card server: %v", err)
		}
	}()

	router := api.NewRouter(store, hub)
	httpServer := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	go func() {
		log.Printf("http server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start http server: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rfidSrv.Close()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Println(err)
	}
}
