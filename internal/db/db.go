package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/behshadrhp/Shadina/internal/consts"
)

func NewDB(dbDriver, addr string, maxOpenConns, maxIdleConns int, maxIdleTimeout string) (*sql.DB, error) {
	db, err := sql.Open(dbDriver, addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxIdleTimeout)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), consts.CTX_TIMEOUT)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
