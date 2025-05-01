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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalw(err.Error(), "event", "load config")
	}

	db := storage.ConnectToDB(cfg.DBConnParams)
	defer db.Close()

	urlStorage, err := storage.NewURLStorage(&db, cfg)
	if err != nil {
		logger.Fatalw(err.Error(), "event", "create storage")
	}
	defer urlStorage.Finalize()

	urlService := service.NewURLService(urlStorage, cfg)
	handler := handler.NewURLHandler(urlService, cfg)

	logger.Infof("Server started on %s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)

	handleChain := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.Authorize(middleware.Compress(middleware.Log(h, logger)), logger, cfg.SecretKey)
	}

	handlePostString := handleChain(handler.ProcessPostURLString)
	handlePostObject := handleChain(handler.ProcessPostURLObject)
	handlePostBatch := handleChain(handler.ProcessPostURLBatch)
	handleGet := handleChain(handler.ProcessGet)
	handleGetSummary := handleChain(handler.ProcessGetSummary)
	handleDeleteUrls := handleChain(handler.ProcessDeleteUrls)
	handlePing := handler.ProcessPing(db)

	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Post("/", handlePostString)
		router.Post("/api/shorten", handlePostObject)
		router.Post("/api/shorten/batch", handlePostBatch)
		router.Get("/api/user/urls", handleGetSummary)
		router.Get("/ping", handlePing)
		router.Get("/{URL}", handleGet)
		router.Delete("/api/user/urls", handleDeleteUrls)
	})

	params := fmt.Sprintf("%s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)
	if err := http.ListenAndServe(params, router); err != nil {
		logger.Fatalw(err.Error(), "event", "start server")
	}
}
