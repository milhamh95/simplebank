package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/milhamh95/simplebank/db/fake"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pkg/password"
	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *fake.FakeStore)
		checkResponse func(rec *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *fake.FakeStore) {
				store.CreateUserReturns(user, nil)
			},
			checkResponse: func(rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchUser(t, rec.Body, user)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeStore := &fake.FakeStore{}
			tc.buildStubs(fakeStore)

			server := newTestServer(t, fakeStore)
			rec := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(rec, req)
			tc.checkResponse(rec)
		})
	}
}

func randomUser(t *testing.T) (user db.User, pwd string) {
	pwd = random.RandomString(6)
	hashedPassword, err := password.HashPassword(pwd)
	require.NoError(t, err)

	user = db.User{
		Username:       random.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       random.RandomOwner(),
		Email:          random.RandomEmail(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
