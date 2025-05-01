package service

import (
	"context"

	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/storage"
	"github.com/rvkarpov/url_shortener/internal/urlutils"
)

type URLService struct {
	urlStorage storage.URLStorage
	cfg        *config.Config
}

func NewURLService(urlStorage storage.URLStorage, cfg *config.Config) *URLService {
	return &URLService{urlStorage: urlStorage, cfg: cfg}
}

func (service *URLService) BeginBatchProcessing(ctx context.Context) error {
	return service.urlStorage.BeginTransaction(ctx)
}

func (service *URLService) EndBatchProcessing(ctx context.Context) error {
	return service.urlStorage.EndTransaction(ctx)
}

func (service *URLService) ProcessLongURL(ctx context.Context, longURL string) (string, error) {
	shortURL := urlutils.GenerateShortURL(longURL, service.cfg.ShortURLLen)
	err := service.urlStorage.StoreURL(ctx, shortURL, longURL)
	return shortURL, err
}

func (service *URLService) ProcessShortURL(ctx context.Context, shortURL string) (string, error) {
	return service.urlStorage.TryGetLongURL(ctx, shortURL)
}

func (service *URLService) GetSummary(ctx context.Context) string {
	return service.urlStorage.GetSummary(ctx)
}
