package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pkg/config"
	"github.com/milhamh95/simplebank/token"
)

type Server struct {
	cfg        config.Config
	store      db.Store
	tokenMaker token.Tokener
	router     *gin.Engine
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

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()

	return server, nil
}

func (s *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)

	authRoutes := router.Group("/").
		Use(
			authMiddleware(s.tokenMaker),
		)

	authRoutes.POST("/accounts", s.createAccount)
	authRoutes.GET("/accounts/:id", s.getAccount)
	authRoutes.GET("/accounts", s.listAccount)

	authRoutes.POST("/transfers", s.createAccount)

	s.router = router
}

func errResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
