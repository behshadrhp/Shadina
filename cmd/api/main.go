package main

import (
	"log"

	"github.com/behshadrhp/Shadina/internal/db"
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
		addr:        env.GetString("SERVER_ADDR", ":8000"),
		version:     env.GetString("APP_VERSION", "0.0.1"),
		environment: env.GetString("APP_ENVIRONMENT", "development"),
		db: dbConfig{
			addr:           env.GetString("DATABASE_ADDR", "postgres://postgres:postgres@127.0.0.1:5435/shadina?sslmode=disable"),
			maxOpenConns:   env.GetInt("MAX_OPEN_CONNS", 30),
			maxIdleConns:   env.GetInt("MAX_IDLE_CONNS", 30),
			maxIdleTimeout: env.GetString("MAX_IDLE_TIMEOUT", "15m"),
		},
	}

	dbDriver := env.GetString("DATABASE_DRIVER", "postgres")
	db, err := db.NewDB(
		dbDriver,
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTimeout,
	)
	if err != nil {
		log.Fatalln(dbDriver, err)
	}
	defer db.Close()
	log.Printf("database(%v) connected pool stablish", dbDriver)

	stg := storage.NewPostgresql(db)

	app := &application{
		config:  cfg,
		storage: stg,
	}

	mux := app.mount()

	log.Fatal(app.run(&mux))
}
