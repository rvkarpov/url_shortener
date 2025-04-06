package mocks

import (
	"context"
	"errors"

	"github.com/rvkarpov/url_shortener/internal/storage"
)

type MockStorage interface {
	storage.URLStorage
	AddTestData(shortURL, longURL string)
}

// implementation temporarily coincides with storage.Storage

type Mock struct {
	urls map[string]string
}

func (m *Mock) StoreURL(ctx context.Context, shortURL, longURL string) error {
	m.urls[shortURL] = longURL
	return nil
}

func (m *Mock) TryGetLongURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, exists := m.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}

	return originalURL, nil
}

func (m *Mock) Finalize() {
}

func (m *Mock) BeginTransaction(ctx context.Context) error {
	return nil
}

func (m *Mock) EndTransaction(ctx context.Context) error {
	return nil
}

func (m *Mock) AddTestData(shortURL, longURL string) {
	m.urls[shortURL] = longURL
}

func NewStorageMock() *Mock {
	return &Mock{urls: make(map[string]string)}
}
