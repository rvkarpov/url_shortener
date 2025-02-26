package handler

import (
	"fmt"
	"net/http"
	"url_shortener/internal/service"
	"url_shortener/internal/url/utils"
)

type UrlHandler struct {
	urlService *service.UrlService
}

func NewUrlHandler(urlService_ *service.UrlService) *UrlHandler {
	return &UrlHandler{urlService: urlService_}
}

func (handler *UrlHandler) ServeHTTP(rsp http.ResponseWriter, rqs *http.Request) {
	switch rqs.Method {
	case http.MethodPost:
		handler.processPost(rsp, rqs)
	case http.MethodGet:
		handler.processGet(rsp, rqs)
	default:
		http.Error(rsp, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *UrlHandler) processPost(rsp http.ResponseWriter, rqs *http.Request) {
	recvUrl, err := utils.TryGetUrlFromPostRqs(rqs)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := handler.urlService.ProcessLongURL(recvUrl)
	rsp.WriteHeader(http.StatusCreated)
	rsp.Write([]byte(shortURL + "\n"))
}

func (handler *UrlHandler) processGet(rsp http.ResponseWriter, rqs *http.Request) {
	recvUrl, err := utils.TryGetUrlFromGetRqs(rqs)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	longURL, err := handler.urlService.ProcessShortURL(recvUrl)
	if err != nil {
		fmt.Printf("Error processing short URL in handler")
		http.Error(rsp, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Printf("Correct!")
	rsp.Header().Set("Location", longURL)
	rsp.WriteHeader(http.StatusTemporaryRedirect)
}
