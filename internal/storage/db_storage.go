package storage

import (
	"context"
	"database/sql"
	"encoding/json"
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
	state     *DBState
	cfg       *config.Config
	deleteCmd *DeleteCmd
}

func (storage *DBStorage) StoreURL(ctx context.Context, shortURL, longURL string) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (userID, longURL, shortURL) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (shortURL) 
		DO NOTHING 
		RETURNING id;`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	var result sql.Result
	if storage.state.Tx != nil {
		result, err = storage.state.Tx.ExecContext(
			ctx,
			query,
			userID,
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
			userID,
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

func (storage *DBStorage) TryGetLongURL(ctx context.Context, shortURL string) (string, bool, error) {
	query := fmt.Sprintf(
		`SELECT longURL, deletedFlag FROM %s WHERE shortURL = $1 LIMIT 1`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	var longURL string
	var deleted bool
	err := storage.state.DB.QueryRowContext(
		ctx,
		query,
		shortURL,
	).Scan(&longURL, &deleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", deleted, fmt.Errorf("short URL '%s' not found", shortURL)
		}
		return "", deleted, fmt.Errorf("database error: %w", err)
	}

	return longURL, deleted, nil
}

func (storage *DBStorage) MarkAsDeleted(ctx context.Context, shortURLs []string) {
	userID, err := getUserID(ctx)
	if err != nil || userID == "" {
		return
	}

	storage.deleteCmd.Append(userID, shortURLs)
}

func (storage *DBStorage) Finalize() {
	storage.deleteCmd.Finalize()
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

func (storage *DBStorage) GetSummary(ctx context.Context) string {
	userID, err := getUserID(ctx)
	if err != nil {
		return ""
	}

	query := fmt.Sprintf(
		`SELECT shortURL, longURL FROM %s WHERE userID = $1`,
		pq.QuoteIdentifier(storage.cfg.TableName),
	)

	rows, err := storage.state.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return ""
	}

	defer rows.Close()

	urls := make([]UserDataStorageItem, 0)
	for rows.Next() {
		var item UserDataStorageItem
		err = rows.Scan(&item.ShortURL, &item.LongURL)
		if err != nil {
			return ""
		}

		item.ShortURL = fmt.Sprintf("%s/%s", storage.cfg.PublishAddr, item.ShortURL)
		urls = append(urls, item)
	}

	err = rows.Err()
	if err != nil {
		return ""
	}

	if len(urls) == 0 {
		return ""
	}

	result, err := json.Marshal(urls)
	if err != nil {
		return ""
	}

	return string(result)
}

func NewDBStorage(state *DBState, cfg *config.Config) (*DBStorage, error) {
	if state.DB == nil {
		return nil, errors.New("database is not available")
	}

	createTable := `
		CREATE TABLE IF NOT EXISTS %s ( 
		id SERIAL PRIMARY KEY,
		userID TEXT NOT NULL,
		longURL TEXT UNIQUE NOT NULL, 
		shortURL VARCHAR(%d) UNIQUE NOT NULL,
		deletedFlag BOOLEAN NOT NULL DEFAULT FALSE,
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

	return &DBStorage{state: state, cfg: cfg, deleteCmd: NewDeleteCmd(state, cfg)}, nil
}
