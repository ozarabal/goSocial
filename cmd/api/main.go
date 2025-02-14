package main

import (
	"expvar"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/ozarabal/goSocial/internal/auth"
	"github.com/ozarabal/goSocial/internal/db"
	"github.com/ozarabal/goSocial/internal/env"
	"github.com/ozarabal/goSocial/internal/mailer"
	"github.com/ozarabal/goSocial/internal/ratelimiter"
	"github.com/ozarabal/goSocial/internal/store"
	"github.com/ozarabal/goSocial/internal/store/cache"
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

	if env.GetString("ENV", "production") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file not found, using Railway environment variables")
		}
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
		redisCfg: redisConfig{
			addr: env.GetString("REDIS_ADDR", "redis_new:6379"),
			pw:		env.GetString("REDIS_PW", ""),
			db:		env.GetInt("REDIS_DB", 0),
			enable: env.GetBool("REDIS_ENABLE", false),
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame: time.Second * 5,
			Enabled: env.GetBool("RATE_LIMITER_ENABLED", true),
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

	var rdb *redis.Client
	if cfg.redisCfg.enable {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cache connection established")
	
		defer rdb.Close()
	}

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	store := store.NewPostgresStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apikey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret, 
		cfg.auth.token.iss, 
		cfg.auth.token.iss)

	app := &application{
		config: cfg,
		store: store,
		cacheStorage: cacheStorage,
		logger: logger,
		mailer: mailer,
		authenticator: jwtAuthenticator,
		rateLimiter: rateLimiter,
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("gorotines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()
	logger.Fatal(app.run(mux))
}