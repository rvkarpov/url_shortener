package service

import (
	"context"

	"github.com/rvkarpov/url_shortener/internal/storage"
	"github.com/rvkarpov/url_shortener/internal/urlutils"
)

type URLService struct {
	urlStorage storage.URLStorage
}

func NewURLService(urlStorage_ storage.URLStorage) *URLService {
	return &URLService{urlStorage: urlStorage_}
}

func (service *URLService) BeginBatchProcessing(ctx context.Context) error {
	return service.urlStorage.BeginTransaction(ctx)
}

func (service *URLService) EndBatchProcessing(ctx context.Context) error {
	return service.urlStorage.EndTransaction(ctx)
}

func (service *URLService) ProcessLongURL(ctx context.Context, longURL string) (string, error) {
	shortURL := urlutils.GenerateShortURL(longURL)
	err := service.urlStorage.StoreURL(ctx, shortURL, longURL)
	return shortURL, err
}

func (service *URLService) ProcessShortURL(ctx context.Context, shortURL string) (string, error) {
	return service.urlStorage.TryGetLongURL(ctx, shortURL)
}
