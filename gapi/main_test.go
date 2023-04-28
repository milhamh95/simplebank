package gapi

import (
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pkg/config"
	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/milhamh95/simplebank/worker"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	cfg := config.Config{
		TokenSymmetricKey:   random.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(cfg, store, taskDistributor)
	require.NoError(t, err)

	return server
}
