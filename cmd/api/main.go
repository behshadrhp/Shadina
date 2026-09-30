package main

import (
	"os"

	"github.com/behshadrhp/Shadina/internal/db"
	"github.com/behshadrhp/Shadina/internal/env"
	"github.com/behshadrhp/Shadina/internal/storage"
	"github.com/behshadrhp/Shadina/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	utils.Validate = validator.New(validator.WithRequiredStructEnabled())
}

//	@title			Shadina API
//	@description	API for Shadina, a social network
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	// logger

	if err := os.MkdirAll("logs", 0755); err != nil {
		panic(err)
	}

	zapCfg := zap.NewProductionConfig()

	zapCfg.OutputPaths = []string{
		"logs/app.log",
	}
	zapCfg.ErrorOutputPaths = []string{
		"logs/error.log",
	}

	log := zap.Must(zapCfg.Build()).Sugar()
	defer log.Sync()

	// load environment
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file")
	}
	cfg := config{
		addr:        env.GetString("SERVER_ADDR", ":8000"),
		url:         env.GetString("SERVER_URL", "127.0.0.1:8000"),
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
	log.Infow("database stable publish", "driver", dbDriver)

	stg := storage.NewPostgresql(db)

	app := &application{
		config:  cfg,
		storage: stg,
		logger:  log,
	}

	mux := app.mount()

	log.Fatal(app.run(&mux))
}
