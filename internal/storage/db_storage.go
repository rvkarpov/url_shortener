package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/rvkarpov/url_shortener/internal/config"
)

type DBState struct {
	DB      *sql.DB
	Enabled bool
}

func (state *DBState) Close() {
	if state.Enabled {
		state.DB.Close()
	}
}

func ConnectToDB(connParams string) DBState {
	if connParams == "" {
		return DBState{DB: nil, Enabled: false}
	}

	db, err := sql.Open("postgres", connParams)
	if err != nil {
		return DBState{DB: nil, Enabled: false}
	}

	return DBState{DB: db, Enabled: true}
}

type DBStorage struct {
	state *DBState
	cfg   config.Config
}

func (storage *DBStorage) StoreURL(shortURL, longURL string) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (longURL, shortURL) VALUES ($1, $2)`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	_, err := storage.state.DB.ExecContext(
		context.Background(),
		query,
		longURL,
		shortURL,
	)

	if err != nil {
		return fmt.Errorf("failed to insert URL: %w", err)
	}
	return nil
}

func (storage *DBStorage) TryGetLongURL(shortURL string) (string, error) {
	query := fmt.Sprintf(
		`SELECT longURL FROM %s WHERE shortURL = $1 LIMIT 1`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	var longURL string
	err := storage.state.DB.QueryRowContext(
		context.Background(),
		query,
		shortURL,
	).Scan(&longURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("short URL '%s' not found", shortURL)
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	return longURL, nil
}

func (storage *DBStorage) Finalize() {
}

func NewDBStorage(state *DBState, cfg config.Config) (*DBStorage, error) {
	if !state.Enabled {
		return nil, errors.New("database is not available")
	}

	createTable := `
		CREATE TABLE IF NOT EXISTS %s ( 
		id SERIAL PRIMARY KEY, 
		longURL TEXT UNIQUE NOT NULL, 
		shortURL VARCHAR(%d) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);`

	_, err := state.DB.ExecContext(
		context.Background(),
		fmt.Sprintf(
			createTable,
			pq.QuoteIdentifier(cfg.TableName),
			cfg.ShortURLLen,
		),
	)
	if err != nil {
		return nil, err
	}

	return &DBStorage{state: state, cfg: cfg}, nil
}
