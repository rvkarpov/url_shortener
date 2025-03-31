package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/urlutils"
)

type URLHandler struct {
	urlService *service.URLService
	cfg        *config.Config
}

func NewURLHandler(urlService_ *service.URLService, cfg_ *config.Config) *URLHandler {
	return &URLHandler{urlService: urlService_, cfg: cfg_}
}

func (handler *URLHandler) ProcessPostCommon(rsp http.ResponseWriter, rqs *http.Request) {
	recvURL, err := urlutils.TryGetURLFromPostRqs(rqs)
	log.Printf("New POST request with URL: %s", recvURL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := handler.urlService.ProcessLongURL(recvURL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Stored short URL: %s", shortURL)
	rsp.WriteHeader(http.StatusCreated)
	rsp.Write([]byte(fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL)))
}

func (handler *URLHandler) ProcessPostObject(rsp http.ResponseWriter, rqs *http.Request) {
	if rqs.Header.Get("Content-Type") != "application/json" {
		http.Error(rsp, "incorrect content type", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(rqs.Body)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	var origin OriginURLInfo
	if err = json.Unmarshal(buf.Bytes(), &origin); err != nil {
		http.Error(rsp, "invalid json", http.StatusBadRequest)
		return
	}
	if origin.URL == "" {
		http.Error(rsp, "url not specified", http.StatusBadRequest)
		return
	}

	log.Printf("New POST request with URL: %s", origin.URL)

	shortURL, err := handler.urlService.ProcessLongURL(origin.URL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Stored short URL: %s", shortURL)

	short := ShortURLInfo{
		Result: fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL),
	}
	out, err := json.Marshal(short)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
	}

	rsp.Header().Add("Content-Type", "application/json")
	rsp.WriteHeader(http.StatusCreated)
	rsp.Write(out)
}

func (handler *URLHandler) ProcessGet(rsp http.ResponseWriter, rqs *http.Request) {
	recvURL := chi.URLParam(rqs, "URL")
	if len(recvURL) == 0 {
		http.Error(rsp, "a non-empty path is expected", http.StatusBadRequest)
	}

	log.Printf("New GET request with short URL: %s", recvURL)

	longURL, err := handler.urlService.ProcessShortURL(recvURL)
	log.Printf("Found original URL: %s", longURL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	rsp.Header().Set("Location", longURL)
	rsp.WriteHeader(http.StatusTemporaryRedirect)
}
