package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/storage"
	"github.com/rvkarpov/url_shortener/internal/urlutils"
)

type URLHandler struct {
	urlService *service.URLService
	cfg        *config.Config
}

func NewURLHandler(urlService_ *service.URLService, cfg_ *config.Config) *URLHandler {
	return &URLHandler{urlService: urlService_, cfg: cfg_}
}

func (handler *URLHandler) ProcessPostURLString(rsp http.ResponseWriter, rqs *http.Request) {
	ctx := rqs.Context()
	recvURL, err := urlutils.TryGetURLFromPostRqs(rqs)
	log.Printf("New POST request with URL: %s", recvURL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := handler.urlService.ProcessLongURL(ctx, recvURL)
	if err != nil {
		if errors.Is(err, &storage.DuplicateURLError{}) {
			log.Printf("Duplicate URL found: %s", shortURL)
			rsp.WriteHeader(http.StatusConflict)
			rsp.Write([]byte(fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL)))
		} else {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	log.Printf("Stored short URL: %s", shortURL)
	rsp.WriteHeader(http.StatusCreated)
	rsp.Write([]byte(fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL)))
}

func (handler *URLHandler) ProcessPostURLObject(rsp http.ResponseWriter, rqs *http.Request) {
	ctx := rqs.Context()

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

	shortURL, err := handler.urlService.ProcessLongURL(ctx, origin.URL)
	if err != nil {
		if errors.Is(err, &storage.DuplicateURLError{}) {
			log.Printf("Duplicate URL found: %s", shortURL)
			handler.publishURLObject(rsp, shortURL, http.StatusConflict)
		} else {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	log.Printf("Stored short URL: %s", shortURL)
	handler.publishURLObject(rsp, shortURL, http.StatusCreated)
}

func (handler *URLHandler) publishURLObject(rsp http.ResponseWriter, shortURL string, status int) {
	short := ShortURLInfo{
		Result: fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL),
	}
	out, err := json.Marshal(short)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	rsp.Header().Add("Content-Type", "application/json")
	rsp.WriteHeader(status)
	rsp.Write(out)
}

func (handler *URLHandler) ProcessPostURLBatch(rsp http.ResponseWriter, rqs *http.Request) {
	ctx := rqs.Context()

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

	log.Print("New POST request with URL batch")

	var inputBatch []OriginURLBatchItem
	if err = json.Unmarshal(buf.Bytes(), &inputBatch); err != nil {
		http.Error(rsp, "invalid json", http.StatusBadRequest)
		return
	}

	if len(inputBatch) == 0 {
		http.Error(rsp, "empty batch", http.StatusBadRequest)
		return
	}

	err = handler.urlService.BeginBatchProcessing(ctx)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	outputBatch := make([]ShortURLBatchItem, 0, len(inputBatch))

	for _, item := range inputBatch {
		log.Printf("New POST request with URL: %s", item.URL)
		shortURL, err := handler.urlService.ProcessLongURL(ctx, item.URL)
		if err != nil {
			if errors.Is(err, &storage.DuplicateURLError{}) {
				log.Printf("Duplicate URL found: %s", shortURL)

				item := ShortURLBatchItem{
					ID:  item.ID,
					URL: fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL),
				}

				out, err := json.Marshal(item)
				if err != nil {
					http.Error(rsp, err.Error(), http.StatusInternalServerError)
				}

				rsp.Header().Add("Content-Type", "application/json")
				rsp.WriteHeader(http.StatusConflict)
				rsp.Write(out)
			} else {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		outputBatch = append(
			outputBatch,
			ShortURLBatchItem{
				ID:  item.ID,
				URL: fmt.Sprintf("%s/%s", handler.cfg.PublishAddr, shortURL),
			},
		)
	}

	err = handler.urlService.EndBatchProcessing(ctx)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(outputBatch)
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

	longURL, err := handler.urlService.ProcessShortURL(rqs.Context(), recvURL)
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Found original URL: %s", longURL)
	rsp.Header().Set("Location", longURL)
	rsp.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *URLHandler) ProcessGetSummary(rsp http.ResponseWriter, rqs *http.Request) {
	summary := handler.urlService.GetSummary(rqs.Context())

	if summary == "" {
		rsp.WriteHeader(http.StatusNoContent)
		return
	}

	rsp.Header().Add("Content-Type", "application/json")
	rsp.WriteHeader(http.StatusOK)
	rsp.Write([]byte(summary))
}

func (handler *URLHandler) ProcessPing(db storage.DBState) http.HandlerFunc {
	return func(rsp http.ResponseWriter, rqs *http.Request) {
		if db.DB == nil {
			http.Error(rsp, "DB disabled", http.StatusInternalServerError)
			return
		}

		err := db.DB.Ping()
		if err != nil {
			http.Error(rsp, fmt.Sprintf("DB connection error: %v", err), http.StatusInternalServerError)
			return
		}

		rsp.Write([]byte("DB connection OK"))
		rsp.WriteHeader(http.StatusOK)
	}
}
