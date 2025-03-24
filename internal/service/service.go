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

func (service *URLService) ProcessLongURL(longURL string) (string, error) {
	shortURL := urlutils.GenerateShortURL(longURL)
	err := service.urlStorage.StoreURL(shortURL, longURL)
	return shortURL, err
}

func (service *URLService) ProcessShortURL(shortURL string) (string, error) {
	return service.urlStorage.TryGetLongURL(shortURL)
}
