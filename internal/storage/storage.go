package storage

import (
	"github.com/rvkarpov/url_shortener/internal/config"
)

type URLStorage interface {
	StoreURL(shortURL, longURL string) error
	TryGetLongURL(shortURL string) (string, error)
	Finalize()
}

func NewURLStorage(dbState *DBState, cfg config.Config) (URLStorage, error) {
	if dbState.Enabled {
		return NewDBStorage(dbState, cfg)
	}

	return NewFileStorage(cfg.StorageFile)
}
