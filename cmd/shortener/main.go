package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/handler"
	"github.com/rvkarpov/url_shortener/internal/middleware"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/storage"
)

func main() {
	logger_, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger_.Sync()
	logger := logger_.Sugar()

	cfg := config.LoadConfig()

	db := storage.ConnectToDB(cfg.DBConnParams)
	defer db.Close()

	urlStorage, err := storage.NewURLStorage(&db, *cfg)
	if err != nil {
		logger.Fatalw(err.Error(), "event", "create storage")
	}
	defer urlStorage.Finalize()

	urlService := service.NewURLService(urlStorage, cfg)
	handler := handler.NewURLHandler(urlService, cfg)

	logger.Infof("Server started on %s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)

	handlePostString := middleware.Compress(middleware.Log(handler.ProcessPostURLString, logger))
	handlePostObject := middleware.Compress(middleware.Log(handler.ProcessPostURLObject, logger))
	handlePostBatch := middleware.Compress(middleware.Log(handler.ProcessPostURLBatch, logger))
	handleGet := middleware.Compress(middleware.Log(handler.ProcessGet, logger))
	handlePing := handler.ProcessPing(db)

	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Post("/", handlePostString)
		router.Post("/api/shorten", handlePostObject)
		router.Post("/api/shorten/batch", handlePostBatch)
		router.Get("/ping", handlePing)
		router.Get("/{URL}", handleGet)
	})

	params := fmt.Sprintf("%s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)
	if err := http.ListenAndServe(params, router); err != nil {
		logger.Fatalw(err.Error(), "event", "start server")
	}
}
