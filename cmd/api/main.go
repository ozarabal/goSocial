package main

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/ozarabal/goSocial/internal/auth"
	"github.com/ozarabal/goSocial/internal/db"
	"github.com/ozarabal/goSocial/internal/env"
	"github.com/ozarabal/goSocial/internal/mailer"
	"github.com/ozarabal/goSocial/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GoSocial API
//	@description	API for GoSocial, a social network for goUser
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
// @description

func main(){

	if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
	dsn := "postgres://admin:adminpassword@127.0.0.1:5432/social?sslmode=disable&client_encoding=UTF8"
	cfg := config{
		addr: env.GetString("ADDR", ":3000"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:3000"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db : dbConfig{
			addr: env.GetString("DB_ADDR", dsn),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "production"),
		mail: mailConfig{
			exp: time.Hour *24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "r.ardhinto@gmail.com"),
			sendGrid: sendGridConfig{
				apikey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER","admin"),
				pass: env.GetString("AUTH_BASIC_PASS","admin"),
			},
			token : tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:	time.Hour * 24 * 3,
				iss:	"goSocial",
			},
		},
		
	}

	//
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	fmt.Printf("Using connection string: %s\n", cfg.db.addr)
	db, err:= db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil{
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	store := store.NewPostgresStorage(db)

	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apikey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret, 
		cfg.auth.token.iss, 
		cfg.auth.token.iss)

	app := &application{
		config: cfg,
		store: store,
		logger: logger,
		mailer: mailer,
		authenticator: jwtAuthenticator,
	}
	mux := app.mount()
	logger.Fatal(app.run(mux))
}