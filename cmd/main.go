package main

import (
	"avito/internal/app"
	"avito/internal/app/routes"
	"avito/internal/cache"
	"avito/internal/config"
	"avito/internal/logger"
	"avito/internal/storage"
	"avito/internal/utils"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	lfu := cache.NewLFUCache(10)

	err = logger.InitLogger()
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	db := storage.NewStorage(storage.GetDatabaseDSN(*cfg))
	defer db.Tm.DB.Close()

	err = utils.GooseUp(db.Tm.DB)
	if err != nil {
		log.Fatalf("migration error: %v", err)
	}

	A := app.NewApp(db, lfu)
	h := routes.NewHandler(*A)

	srv := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: h,
	}

	done := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)
		<-sigs
		if err = srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(done)
	}()

	err = srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
