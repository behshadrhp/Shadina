package main

import (
	"log"

	"github.com/behshadrhp/Shadina/internal/env"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/joho/godotenv"
)

func init() {
	log.SetPrefix("Shadina: ")

	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file")
	}
}

func main() {
	cfg := config{
		addr: env.GetString("SERVER_ADDR", ":8000"),
		db: dbConfig{
			addr:           env.GetString("DATABASE_ADDR", "postgres://postgres:postgres@localhost/shadina?sslmode=disable"),
			maxOpenConns:   env.GetInt("MAX_OPEN_CONNS", 30),
			maxIdleConns:   env.GetInt("MAX_IDLE_CONNS", 30),
			maxIdleTimeout: env.GetString("MAX_IDLE_TIMEOUT", "15m"),
		},
	}

	stg := storage.NewPostgresql(nil)

	app := &application{
		config:  cfg,
		storage: stg,
	}

	mux := app.mount()

	log.Fatal(app.run(&mux))
}
