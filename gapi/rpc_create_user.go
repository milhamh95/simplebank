package gapi

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/milhamh95/simplebank/worker"
	"time"

	"github.com/lib/pq"
	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pb"
	"github.com/milhamh95/simplebank/pkg/password"
	"github.com/milhamh95/simplebank/pkg/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := password.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"hash password: %s", err.Error(),
		)
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(
					codes.AlreadyExists,
					"username already exists: %s", err.Error(),
				)
			}
		}

		return nil, status.Errorf(codes.Internal, "create user: %s", err.Error())
	}

	// TODO: use db transaction
	taskPayload := &worker.PayloadSendVerifyEmail{
		Username: user.Username,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	err = s.taskDistributor.DistributeTaskSenderVerifyEmail(ctx, taskPayload, opts...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "distribute task to send verify email: %s", err.Error())
	}

	rsp := &pb.CreateUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreatedAt),
		},
	}

	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	err := validator.ValidateUsername(req.GetUsername())
	if err != nil {
		violations = append(
			violations,
			fieldViolation("username", err),
		)
	}

	err = validator.ValidatePassword(req.GetPassword())
	if err != nil {
		violations = append(
			violations,
			fieldViolation("password", err),
		)
	}

	err = validator.ValidateFullName(req.GetFullName())
	if err != nil {
		violations = append(
			violations,
			fieldViolation("full_name", err),
		)
	}

	err = validator.ValidateEmail(req.GetEmail())
	if err != nil {
		violations = append(
			violations,
			fieldViolation("email", err),
		)
	}

	return violations
}
