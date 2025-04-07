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
	DB *sql.DB
	Tx *sql.Tx
}

type DuplicateError struct {
	Err error
}

func (state *DBState) Close() {
	if state.DB != nil {
		state.DB.Close()
	}
}

func ConnectToDB(connParams string) DBState {
	if connParams == "" {
		return DBState{DB: nil, Tx: nil}
	}

	db, err := sql.Open("postgres", connParams)
	if err != nil {
		return DBState{DB: nil, Tx: nil}
	}

	return DBState{DB: db, Tx: nil}
}

type DBStorage struct {
	state *DBState
	cfg   config.Config
}

func (storage *DBStorage) StoreURL(ctx context.Context, shortURL, longURL string) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (longURL, shortURL) 
		VALUES ($1, $2) 
		ON CONFLICT (shortURL) 
		DO NOTHING 
		RETURNING id;`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	var err error
	var result sql.Result

	if storage.state.Tx != nil {
		result, err = storage.state.Tx.ExecContext(
			ctx,
			query,
			longURL,
			shortURL,
		)

		if err != nil {
			if rollbackErr := storage.state.Tx.Rollback(); rollbackErr != nil {
				storage.state.Tx = nil
				return fmt.Errorf("failed to insert URL: %v, rollback: %v", err, rollbackErr)
			}
		}

	} else {
		result, err = storage.state.DB.ExecContext(
			ctx,
			query,
			longURL,
			shortURL,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to insert URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return NewDuplicateURLError(shortURL)
	}

	return nil
}

func (storage *DBStorage) TryGetLongURL(ctx context.Context, shortURL string) (string, error) {
	query := fmt.Sprintf(
		`SELECT longURL FROM %s WHERE shortURL = $1 LIMIT 1`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	var longURL string
	err := storage.state.DB.QueryRowContext(
		ctx,
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

func (storage *DBStorage) BeginTransaction(ctx context.Context) error {
	tx, err := storage.state.DB.Begin()
	if err != nil {
		return err
	}

	storage.state.Tx = tx
	return nil
}

func (storage *DBStorage) EndTransaction(ctx context.Context) error {
	if storage.state.Tx == nil {
		return nil
	}

	err := storage.state.Tx.Commit()
	if err != nil {
		if rollbackErr := storage.state.Tx.Rollback(); rollbackErr != nil {
			storage.state.Tx = nil
			return fmt.Errorf("commit failed: %v, rollback: %v", err, rollbackErr)
		}
	}

	storage.state.Tx = nil
	return nil
}

func NewDBStorage(state *DBState, cfg config.Config) (*DBStorage, error) {
	if state.DB == nil {
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
