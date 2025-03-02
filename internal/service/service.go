package service

import (
	"github.com/rvkarpov/url_shortener/internal/storage"
	"github.com/rvkarpov/url_shortener/internal/urlutils"
)

type URLService struct {
	urlStorage storage.URLStorage
}

func NewURLService(urlStorage_ storage.URLStorage) *URLService {
	return &URLService{urlStorage: urlStorage_}
}

func (service *URLService) ProcessLongURL(longURL string) string {
	shortURL := urlutils.GenerateShortURL(longURL)
	service.urlStorage.StoreURL(shortURL, longURL)
	return shortURL
}

func (service *URLService) ProcessShortURL(shortURL string) (string, error) {
	return service.urlStorage.TryGetLongURL(shortURL)
}
