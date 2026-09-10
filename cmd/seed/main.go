package main

import (
	"log"

	"github.com/behshadrhp/Shadina/internal/db"
	"github.com/behshadrhp/Shadina/internal/env"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file")
	}

	addr := env.GetString("DATABASE_ADDR", "postgres://postgres:postgres@127.0.0.1:5435/shadina?sslmode=disable")
	maxOpenConns := env.GetInt("MAX_OPEN_CONNS", 30)
	maxIdleConns := env.GetInt("MAX_IDLE_CONNS", 30)
	maxIdleTimeout := env.GetString("MAX_IDLE_TIMEOUT", "15m")

	dbDriver := env.GetString("DB_DRIVER", "postgres")
	database, err := db.NewDB(
		dbDriver,
		addr,
		maxOpenConns,
		maxIdleConns,
		maxIdleTimeout,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer database.Close()

	store := storage.NewPostgresql(database)

	db.Seed(store, database)
}
