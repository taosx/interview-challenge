package storage

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

type SQLiteStorage struct {
	*sqlx.DB
}

func NewStorageSQLite(dsn string) *SQLiteStorage {
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	return &SQLiteStorage{db}
}
