package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"yube/internal/db"
	"yube/internal/feed"
	"yube/internal/web"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dbPath := getenv("DB_PATH", "yube.db")
	addr := getenv("ADDR", ":8080")

	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	refresher := feed.NewRefresher(store)
	settings, err := store.GetSettings(ctx)
	if err != nil {
		log.Fatal(err)
	}
	refresher.ApplySettings(settings)
	refresher.Start(ctx)

	srv := &http.Server{Addr: addr, Handler: web.New(store, refresher).Routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
