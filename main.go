package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/hibiken/asynq"
	"github.com/milhamh95/simplebank/mail"
	"github.com/milhamh95/simplebank/worker"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog/log"
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
		log.Fatal().
			Err(err).
			Msg("cannot load config")
	}

	if cfg.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("cannot connect to db")
	}

	runDBMigration(cfg.MigrationURL, cfg.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	// use goroutine for task processor
	// because it will block by get all data
	// from redis
	go runTaskProcessor(cfg, redisOpt, store)
	go runGatewayServer(cfg, store, taskDistributor)
	runGrpcServer(cfg, store, taskDistributor)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("cannot create new migrate instance")
	}

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(cfg config.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(cfg.EmailSenderName, cfg.EmailSenderAddress, cfg.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("start task processor")
	}
}

func runGrpcServer(cfg config.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(cfg, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("initialize server:")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)

	// allow gRPC client explore rpc in server
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	log.Printf("start GRPC server at: %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gRPC server")
	}
}

func runGatewayServer(cfg config.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(cfg, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("initialize server")
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
		log.Fatal().Err(err).Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))

	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	log.Printf("start HTTP Gateway server at: %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server")
	}
}

func runGinServer(cfg config.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Err(err).Msg("initialize server")
	}

	err = server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("can't start server")
	}
}
