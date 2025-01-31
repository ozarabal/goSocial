package main

import (
	"log"

	"github.com/ozarabal/goSocial/internal/db"
	"github.com/ozarabal/goSocial/internal/env"
	"github.com/ozarabal/goSocial/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@db:5432/social?sslmode=disable&client_encoding=UTF8")
	conn, err := db.New(addr, 3,3,"15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	store := store.NewPostgresStorage(conn)

	db.Seed(store,conn)
}