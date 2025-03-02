package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/handler"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/storage"
)

func main() {
	cfg := config.LoadConfig()
	urlStorage := storage.NewStorage()
	urlService := service.NewURLService(urlStorage)
	handler := handler.NewURLHandler(urlService, &cfg)

	log.Printf("Server started on %s:%d\n", cfg.Host, cfg.Port)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.ProcessRqs)
	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), mux)
}
