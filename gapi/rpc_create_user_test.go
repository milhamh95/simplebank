package gapi

import (
	"context"
	"database/sql"
	"github.com/milhamh95/simplebank/db/fake"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pb"
	"github.com/milhamh95/simplebank/pkg/password"
	"github.com/milhamh95/simplebank/pkg/random"
	workerFake "github.com/milhamh95/simplebank/worker/fake"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestCreateUserAPI(t *testing.T) {
	user, pwsd := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *fake.FakeStore, taskDistributor *workerFake.FakeTaskDistributor)
		checkResponse func(t *testing.T, resp *pb.CreateUserResponse, err error)
	}{
		{
			name: "success",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: pwsd,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *fake.FakeStore, taskDistributor *workerFake.FakeTaskDistributor) {
				// 1
				//taskDistributor.DistributeTaskSenderVerifyEmailReturns(errors.New("err call"))

				//store.CreateUserTrxReturns(db.CreateUserTxResult{User: user}, nil)

				// 2
				//taskDistributor.DistributeTaskSenderVerifyEmailStub = func(ctx context.Context, email *worker.PayloadSendVerifyEmail, option ...asynq.Option) error {
				//	return nil
				//}

				store.CreateUserTrxStub = func(ctx context.Context, params db.CreateUserTxParams) (db.CreateUserTxResult, error) {
					return db.CreateUserTxResult{User: user}, nil
				}
			},
			checkResponse: func(t *testing.T, resp *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)

				createdUser := resp.GetUser()
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.FullName, createdUser.FullName)
				require.Equal(t, user.Email, createdUser.Email)
			},
		},
		{
			name: "internal error",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: pwsd,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *fake.FakeStore, taskDistributor *workerFake.FakeTaskDistributor) {
				store.CreateUserTrxStub = func(ctx context.Context, params db.CreateUserTxParams) (db.CreateUserTxResult, error) {
					return db.CreateUserTxResult{}, sql.ErrConnDone
				}
			},
			checkResponse: func(t *testing.T, resp *pb.CreateUserResponse, err error) {
				require.Error(t, err)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeStore := &fake.FakeStore{}
			taskDistributor := &workerFake.FakeTaskDistributor{}
			tc.buildStubs(fakeStore, taskDistributor)

			server := newTestServer(t, fakeStore, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
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
