package repo

import (
	"database/sql"

	"im/pkg/db"
)

func NewPool(dsn string) (*sql.DB, error) {
	return db.NewDB(dsn)
}
