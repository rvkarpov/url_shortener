package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rvkarpov/url_shortener/internal/config"
)

func getUserID(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDKey{Name: "userID"})
	if val == nil {
		return "", fmt.Errorf("no userID value in context")
	}

	userID, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("incorrect userID type")
	}

	return userID, nil
}

type UserIDKey struct {
	Name string
}

type UserDataStorageItem struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

type UserDataStorage struct {
	urls map[string][]UserDataStorageItem
	cfg  *config.Config
}

func NewUserDataStorage(cfg *config.Config) *UserDataStorage {
	return &UserDataStorage{urls: make(map[string][]UserDataStorageItem), cfg: cfg}
}

func (storage *UserDataStorage) append(userID, longURL, shortURL string) {
	if userID == "" {
		return
	}

	urls := storage.urls[userID]

	shortURLnorm := fmt.Sprintf("%s/%s", storage.cfg.PublishAddr, shortURL)
	urls = append(urls, UserDataStorageItem{ShortURL: shortURLnorm, LongURL: longURL})
	storage.urls[userID] = urls
}

func (storage *UserDataStorage) getSummary(ctx context.Context) string {
	userID, err := getUserID(ctx)
	if err != nil {
		return ""
	}

	urls := storage.urls[userID]
	if len(urls) == 0 {
		return ""
	}

	out, err := json.Marshal(urls)
	if err != nil {
		return ""
	}

	return string(out)
}
