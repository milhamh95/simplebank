package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/milhamh95/simplebank/token"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	req *http.Request,
	tokenMaker token.Tokener,
	authorizationType, username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authHeader := fmt.Sprintf("%s %s", authorizationType, token)
	req.Header.Set(authorizationHeaderKey, authHeader)
}

func TestAuthMiddlewre(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Tokener)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					authorizationTypeBearer,
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name:      "no authorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "unsupported authorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					"unsupported",
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "invalid authorization format",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					"",
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "expired token",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					"",
					"user",
					-time.Minute,
				)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}
