package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pkg/config"
	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	cfg := config.Config{
		TokenSymmetricKey:   random.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(cfg, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
