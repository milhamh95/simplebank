package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/milhamh95/simplebank/db/fake"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/milhamh95/simplebank/token"

	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	testCases := []struct {
		name               string
		accountID          int64
		setupAuth          func(t *testing.T, req *http.Request, tokenMaker token.Tokener)
		callGetAccountStub bool
		getAccountStub     func(store *fake.FakeStore)
		checkResponse      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "success",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					time.Minute,
				)
			},
			callGetAccountStub: true,
			getAccountStub: func(store *fake.FakeStore) {
				store.GetAccountReturns(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "not found",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					time.Minute,
				)
			},
			callGetAccountStub: true,
			getAccountStub: func(store *fake.FakeStore) {
				store.GetAccountReturns(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "unknown error",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					time.Minute,
				)
			},
			callGetAccountStub: true,
			getAccountStub: func(store *fake.FakeStore) {
				store.GetAccountReturns(db.Account{}, errors.New("unknown error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "invalid id",
			accountID: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Tokener) {
				addAuthorization(
					t,
					req,
					tokenMaker,
					authorizationTypeBearer,
					user.Username,
					time.Minute,
				)
			},
			callGetAccountStub: false,
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeStore := &fake.FakeStore{}

			if tc.callGetAccountStub {
				tc.getAccountStub(fakeStore)
			}

			server := newTestServer(t, fakeStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       random.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  random.RandomMoney(),
		Currency: random.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
