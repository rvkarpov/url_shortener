package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

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

func (handler *URLHandler) ProcessRqs(rsp http.ResponseWriter, rqs *http.Request) {
	switch rqs.Method {
	case http.MethodPost:
		handler.processPost(rsp, rqs)
	case http.MethodGet:
		handler.processGet(rsp, rqs)
	default:
		http.Error(rsp, "Method not allowed", http.StatusBadRequest)
	}
}

func (handler *URLHandler) processPost(rsp http.ResponseWriter, rqs *http.Request) {
	recvURL, err := urlutils.TryGetURLFromPostRqs(rqs)
	log.Printf("New POST request with URL: %s", recvURL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := handler.urlService.ProcessLongURL(recvURL)
	log.Printf("Stored short URL: %s", shortURL)
	rsp.WriteHeader(http.StatusCreated)

	result := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", handler.cfg.Host, handler.cfg.Port),
		Path:   "/" + shortURL,
	}

	rsp.Write([]byte(result.String()))
}

func (handler *URLHandler) processGet(rsp http.ResponseWriter, rqs *http.Request) {
	log.Println("New GET request")
	recvURL, err := urlutils.TryGetURLFromGetRqs(rqs)

	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
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
