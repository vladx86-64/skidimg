package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
)

type Database struct {
	db *sqlx.DB
}

func NewDatabase() (*Database, error) {
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is not set")
	}

	// dsn := "user=postgres password=carln109 dbname=skidimg sslmode=disable host=localhost port=5432"
	// dsn := "user=postgres password=carln109 dbname=skidimg sslmode=disable host=localhost port=5433"
	dsn := fmt.Sprintf("user=postgres password=%s dbname=skidimg sslmode=disable host=skidimg-postgres port=5432", password)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) GetDB() *sqlx.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}
