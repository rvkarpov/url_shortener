package service

import (
	"url_shortener/internal/url/storage"
	"url_shortener/internal/url/utils"
)

type UrlService struct {
	urlStorage storage.UrlStorage
}

func NewUrlService(urlStorage_ storage.UrlStorage) *UrlService {
	return &UrlService{urlStorage: urlStorage_}
}

func (service *UrlService) ProcessLongURL(longURL string) string {
	shortUrl := utils.GenerateShortURL(longURL)
	service.urlStorage.StoreUrl(shortUrl, longURL)
	return shortUrl
}

func (service *UrlService) ProcessShortURL(shortURL string) (string, error) {
	return service.urlStorage.TryGetLongURL(shortURL)
}
