package gapi

import (
	"fmt"

	"github.com/gin-gonic/gin"

	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pb"
	"github.com/milhamh95/simplebank/pkg/config"
	"github.com/milhamh95/simplebank/token"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	cfg        config.Config
	store      db.Store
	tokenMaker token.Tokener
}

func NewServer(cfg config.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPaseto(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("init token maker: %w", err)
	}
	server := &Server{
		cfg:        cfg,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

func errResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}