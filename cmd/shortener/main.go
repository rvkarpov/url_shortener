package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/handler"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/storage"
)

func main() {
	cfg := config.LoadConfig()
	urlStorage := storage.NewStorage()
	urlService := service.NewURLService(urlStorage)
	handler := handler.NewURLHandler(urlService, cfg)

	log.Printf("Server started on %s:%d\n", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port)

	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Post("/", handler.ProcessPost)
		router.Get("/{URL}", handler.ProcessGet)
	})

	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.LaunchAddr.Host, cfg.LaunchAddr.Port), router)
}
