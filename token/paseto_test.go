package token

import (
	"testing"
	"time"

	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	pasetoToken, err := NewPaseto(random.RandomString(32))
	require.NoError(t, err)

	username := random.RandomOwner()
	durationTimeMinute := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(durationTimeMinute)

	token, payload, err := pasetoToken.CreateToken(username, durationTimeMinute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = pasetoToken.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	pasetoToken, err := NewPaseto(random.RandomString(32))
	require.NoError(t, err)

	token, payload, err := pasetoToken.CreateToken(random.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = pasetoToken.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
