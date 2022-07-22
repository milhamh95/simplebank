package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/milhamh95/simplebank/api"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pkg/config"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal("initialize server:", err)
	}

	err = server.Start(cfg.ServerAddress)
	if err != nil {
		log.Fatalf("can't start server: %v", err)
	}
}
