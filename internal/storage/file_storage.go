package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"

	"github.com/rvkarpov/url_shortener/internal/config"
)

type FileStorage struct {
	urls     map[string]string
	file     *os.File
	writer   *bufio.Writer
	userData *UserDataStorage
}

func (storage *FileStorage) StoreURL(ctx context.Context, shortURL, longURL string) error {
	_, exists := storage.urls[shortURL]
	if exists {
		return NewDuplicateURLError(shortURL)
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	storage.urls[shortURL] = longURL
	if err := storage.writeItem(userID, shortURL, longURL); err != nil {
		return err
	}

	storage.userData.append(userID, longURL, shortURL)
	return nil
}

func (storage *FileStorage) TryGetLongURL(ctx context.Context, shortURL string) (string, bool, error) {
	originalURL, exists := storage.urls[shortURL]
	deleted := false
	if !exists {
		return "", deleted, errors.New("URL not found")
	}

	return originalURL, deleted, nil
}

func (storage *FileStorage) MarkAsDeleted(ctx context.Context, shortURLs []string) {
}

func (storage *FileStorage) Finalize() {
	storage.writer.Flush()
	storage.file.Close()
}

type StorageItem struct {
	ItemID      string `json:"item_id"`
	UserID      string `json:"user_id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (storage *FileStorage) writeItem(userID, shortURL, longURL string) error {
	item := StorageItem{
		ItemID:      strconv.Itoa(len(storage.urls)),
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: longURL,
	}

	data, err := json.Marshal(&item)
	if err != nil {
		return err
	}

	if _, err := storage.writer.Write(data); err != nil {
		return err
	}

	if err := storage.writer.WriteByte('\n'); err != nil {
		return err
	}

	return storage.writer.Flush()
}

func (storage *FileStorage) BeginTransaction(ctx context.Context) error {
	return nil
}

func (storage *FileStorage) EndTransaction(ctx context.Context) error {
	return nil
}

func (storage *FileStorage) GetSummary(ctx context.Context) string {
	return storage.userData.getSummary(ctx)
}

func NewFileStorage(cfg *config.Config) (*FileStorage, error) {
	file, err := os.OpenFile(cfg.StorageFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	urls := make(map[string]string)
	userData := NewUserDataStorage(cfg)

	decoder := json.NewDecoder(file)
	var item StorageItem
	for {
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		urls[item.ShortURL] = item.OriginalURL
		userData.append(item.UserID, item.OriginalURL, item.ShortURL)
	}

	return &FileStorage{urls: urls, file: file, writer: bufio.NewWriter(file), userData: userData}, nil
}
