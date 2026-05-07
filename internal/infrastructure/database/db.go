package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	Conn *sqlx.DB
}

func dsn() string {
	DBUser := os.Getenv("DB_USER")
	DBPass := os.Getenv("DB_PASSWORD")
	DBHost := os.Getenv("DB_HOST")
	DBPort := os.Getenv("DB_PORT")
	DBName := os.Getenv("DB_NAME")
	DBSSLMode := os.Getenv("DB_SSLMODE")
	if DBSSLMode == "" {
		DBSSLMode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		DBUser, DBPass, DBHost, DBPort, DBName, DBSSLMode,
	)
}

func NewDb() *Database {
	conn, err := sqlx.Connect("postgres", dsn())
	if err != nil {
		log.Fatal(err.Error())
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &Database{
		Conn: conn,
	}
}
