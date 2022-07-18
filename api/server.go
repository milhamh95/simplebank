package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/milhamh95/simplebank/db/sqlc"
)

type Server struct {
	store  *db.Store
	rotuer *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	server.rotuer = router
	return server
}

func errResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (s *Server) Start(addr string) error {
	return s.rotuer.Run(addr)
}
