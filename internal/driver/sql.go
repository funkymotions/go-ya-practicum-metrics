package driver

import (
	"context"
	"database/sql"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/db"
	_ "github.com/lib/pq"
)

type SQLDriver struct {
	DB *sql.DB
}

func NewSQLDriver(c *db.DBConfig) (*SQLDriver, error) {
	db, err := sql.Open(c.Type, c.DSN)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &SQLDriver{DB: db}, nil
}
