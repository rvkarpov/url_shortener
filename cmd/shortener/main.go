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
	urlStorage := storage.NewStorage()
	urlService := service.NewURLService(urlStorage)
	handler := handler.NewURLHandler(urlService, cfg)

	logger.Infof("Server started on %s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)

	handlePostString := middleware.Log(handler.ProcessPostCommon, logger)
	handlePostObject := middleware.Log(handler.ProcessPostObject, logger)
	handleGet := middleware.Log(handler.ProcessGet, logger)

	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Post("/", handlePostString)
		router.Post("/api/shorten", handlePostObject)
		router.Get("/{URL}", handleGet)
	})

	params := fmt.Sprintf("%s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)
	if err := http.ListenAndServe(params, router); err != nil {
		logger.Fatalw(err.Error(), "event", "start server")
	}
}
