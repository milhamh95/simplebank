package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	jwt, err := NewJWT(random.RandomString(32))
	require.NoError(t, err)

	username := random.RandomOwner()
	durationTimeMinute := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(durationTimeMinute)

	token, payload, err := jwt.CreateToken(username, durationTimeMinute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = jwt.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	jwt, err := NewJWT(random.RandomString(32))
	require.NoError(t, err)

	token, payload, err := jwt.CreateToken(random.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = jwt.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(random.RandomOwner(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	jwtMaker, err := NewJWT(random.RandomString(32))
	require.NoError(t, err)

	payload, err = jwtMaker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
