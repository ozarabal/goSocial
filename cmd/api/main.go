package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/ozarabal/goSocial/internal/db"
	"github.com/ozarabal/goSocial/internal/env"
	"github.com/ozarabal/goSocial/internal/store"
)

const version = "0.0.1"

func main(){
	if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
	dsn := "postgres://admin:adminpassword@127.0.0.1:5432/social?sslmode=disable&client_encoding=UTF8"
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db : dbConfig{
			addr: env.GetString("DB_ADDR", dsn),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime: env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},	
	}

	fmt.Printf("Using connection string: %s\n", cfg.db.addr)
	db, err:= db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil{
		log.Panic(err)
	}
	defer db.Close()
	log.Println("database connection pool established")

	store := store.NewPostgresStorage(db)

	app := &application{
		config: cfg,
		store: store,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}