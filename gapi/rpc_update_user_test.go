package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/milhamh95/simplebank/db/fake"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pb"
	"github.com/milhamh95/simplebank/pkg/random"
	"github.com/milhamh95/simplebank/token"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)

	newName := random.RandomOwner()
	newEmail := random.RandomEmail()

	invalidEmail := "invalid-email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *fake.FakeStore)
		buildContext  func(t *testing.T, tokenMaker token.Tokener) context.Context
		checkResponse func(t *testing.T, resp *pb.UpdateUserResponse, err error)
	}{
		{
			name: "success",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *fake.FakeStore) {

				store.UpdateUserStub = func(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
					return db.User{
						Username:          user.Username,
						HashedPassword:    user.HashedPassword,
						FullName:          newName,
						Email:             newEmail,
						PasswordChangedAt: user.PasswordChangedAt,
						CreatedAt:         user.CreatedAt,
						IsEmailVerified:   user.IsEmailVerified,
					}, nil
				}
			},
			buildContext: func(t *testing.T, tokenMaker token.Tokener) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			checkResponse: func(t *testing.T, resp *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)

				updatedUser := resp.GetUser()
				require.Equal(t, user.Username, updatedUser.Username)
				require.Equal(t, newName, updatedUser.FullName)
				require.Equal(t, newEmail, updatedUser.Email)
			},
		},
		{
			name: "user not found",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *fake.FakeStore) {

				store.UpdateUserStub = func(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
					return db.User{}, sql.ErrNoRows
				}
			},
			buildContext: func(t *testing.T, tokenMaker token.Tokener) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			checkResponse: func(t *testing.T, resp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "expired token",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *fake.FakeStore) {
				return
			},
			buildContext: func(t *testing.T, tokenMaker token.Tokener) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, -time.Minute)
			},
			checkResponse: func(t *testing.T, resp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "invalid email",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &invalidEmail,
			},
			buildStubs: func(store *fake.FakeStore) {
				return
			},
			buildContext: func(t *testing.T, tokenMaker token.Tokener) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			checkResponse: func(t *testing.T, resp *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeStore := &fake.FakeStore{}
			tc.buildStubs(fakeStore)

			server := newTestServer(t, fakeStore, nil)

			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Tokener, user db.User, duration time.Duration) context.Context {
	ctx := context.Background()
	accessToken, _, err := tokenMaker.CreateToken(user.Username, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}
	return metadata.NewIncomingContext(ctx, md)
}
