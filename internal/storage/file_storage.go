package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
)

type Storage struct {
	urls   map[string]string
	file   *os.File
	writer *bufio.Writer
}

func (storage *Storage) StoreURL(shortURL, longURL string) error {
	storage.urls[shortURL] = longURL
	return storage.writeItem(shortURL, longURL)
}

func (storage *Storage) TryGetLongURL(shortURL string) (string, error) {
	originalURL, exists := storage.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}

	return originalURL, nil
}

func (storage *Storage) Finalize() {
	storage.writer.Flush()
	storage.file.Close()
}

type StorageItem struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (storage *Storage) writeItem(shortURL, longURL string) error {
	item := StorageItem{
		UUID:        strconv.Itoa(len(storage.urls)),
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

func NewStorage(filepath string) (*Storage, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	urls := make(map[string]string)
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
	}

	return &Storage{urls: urls, file: file, writer: bufio.NewWriter(file)}, nil
}
