package zerokit

import (
	"database/sql"
	"log"

	"im/pkg/repo"
)

func MustMySQL(dsn string) *sql.DB {
	db, err := repo.NewPool(dsn)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	return db
}
