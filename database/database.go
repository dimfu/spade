package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/dimfu/spade/config"
	_ "github.com/go-sql-driver/mysql"
)

var _db *sql.DB

func GetDB() *sql.DB {
	return _db
}

func Init() *sql.DB {
	cfg := config.GetEnv()
	src := fmt.Sprintf("%s:%s@tcp(db:%s)/%s", cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_PORT, cfg.DB_NAME)
	db, err := sql.Open("mysql", src)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	log.Printf("Established connection to database")

	_db = db

	return db
}
