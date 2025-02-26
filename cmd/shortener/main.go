package main

import (
	"fmt"
	"log"
	"net/http"
	"url_shortener/internal/config"
	"url_shortener/internal/handler"
	"url_shortener/internal/service"
	"url_shortener/internal/url/storage"
)

func main() {
	urlStorage := storage.NewStorage()
	urlService := service.NewUrlService(urlStorage)
	handler := handler.NewUrlHandler(urlService)

	cfg := config.LoadConfig()
	log.Printf("Server started on :%d\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler)
}
