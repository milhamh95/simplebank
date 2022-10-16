package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/milhamh95/simplebank/api"
	db "github.com/milhamh95/simplebank/db/sqlc"
	_ "github.com/milhamh95/simplebank/doc/statik"
	"github.com/milhamh95/simplebank/gapi"
	"github.com/milhamh95/simplebank/pb"
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

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(cfg, store)
	runGrpcServer(cfg, store)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance")
	}

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("failed to run migrate up: ", err)
	}

	log.Println("db migrated successfully")
}

func runGrpcServer(cfg config.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal("initialize server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)

	// allow gRPC client explore rpc in server
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}

	log.Printf("start GRPC server at: %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server")
	}
}

func runGatewayServer(cfg config.Config, store db.Store) {
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		log.Fatal("initialize server:", err)
	}

	grpcMux := runtime.NewServeMux(
		// json option
		// to make response http gateway same as defined in protobuf file
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs:", err)
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))

	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start HTTP Gateway server at: %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:", err)
	}
}

func runGinServer(cfg config.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal("initialize server:", err)
	}

	err = server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatalf("can't start server: %v", err)
	}
}
