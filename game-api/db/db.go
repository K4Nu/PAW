package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	dsn := "postgres://gameuser:gamepass@localhost:5432/game?sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("Cannot ping DB:", err)
	}
	fmt.Println("✅ Connected to Postgres!")
}
