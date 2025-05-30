package storage

import (
	"context"

	"github.com/rvkarpov/url_shortener/internal/config"
)

type URLStorage interface {
	StoreURL(ctx context.Context, shortURL, longURL string) error
	TryGetLongURL(ctx context.Context, shortURL string) (string, bool, error)
	MarkAsDeleted(ctx context.Context, shortURL []string)
	Finalize()

	BeginTransaction(ctx context.Context) error
	EndTransaction(ctx context.Context) error

	GetSummary(ctx context.Context) string
}

func NewURLStorage(dbState *DBState, cfg *config.Config) (URLStorage, error) {
	if dbState.DB != nil {
		return NewDBStorage(dbState, cfg)
	}

	return NewFileStorage(cfg)
}
