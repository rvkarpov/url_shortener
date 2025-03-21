package storage

import "errors"

type URLStorage interface {
	StoreURL(shortURL, longURL string)
	TryGetLongURL(shortURL string) (string, error)
}

type Storage struct {
	urls map[string]string
}

func (storage *Storage) StoreURL(shortURL, longURL string) {
	storage.urls[shortURL] = longURL
}

func (storage *Storage) TryGetLongURL(shortURL string) (string, error) {
	originalURL, exists := storage.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}

	return originalURL, nil
}

func NewStorage() *Storage {
	return &Storage{urls: make(map[string]string)}
}
