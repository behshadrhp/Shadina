package main

import (
	"log"

	"github.com/behshadrhp/Shadina/internal/env"
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
	}

	app := &application{
		config: cfg,
	}

	mux := app.mount()

	log.Fatal(app.run(&mux))
}
