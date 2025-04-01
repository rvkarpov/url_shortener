package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DBstate struct {
	DB      *sql.DB
	Enabled bool
}

func (state *DBstate) Close() {
	if state.Enabled {
		state.DB.Close()
	}
}

func InitDB(connParams string) DBstate {
	if connParams == "" {
		return DBstate{DB: nil, Enabled: false}
	}

	db, err := sql.Open("postgres", connParams)
	if err != nil {
		return DBstate{DB: nil, Enabled: false}
	}

	return DBstate{DB: db, Enabled: true}
}
