package gapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/milhamh95/simplebank/db/sqlc"
	"github.com/milhamh95/simplebank/pb"
	"github.com/milhamh95/simplebank/pkg/password"
	"github.com/milhamh95/simplebank/pkg/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := password.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"hash password: %s", err.Error(),
			)
		}

		arg.HashedPassword = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}

		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := s.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "update user: %s", err.Error())
	}

	rsp := &pb.UpdateUserResponse{
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

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	err := validator.ValidateUsername(req.GetUsername())
	if err != nil {
		violations = append(
			violations,
			fieldViolation("username", err),
		)
	}

	if req.Password != nil {
		err = validator.ValidatePassword(req.GetPassword())
		if err != nil {
			violations = append(
				violations,
				fieldViolation("password", err),
			)
		}
	}

	if req.FullName != nil {
		err = validator.ValidateFullName(req.GetFullName())
		if err != nil {
			violations = append(
				violations,
				fieldViolation("full_name", err),
			)
		}
	}

	if req.Email != nil {
		err = validator.ValidateEmail(req.GetEmail())
		if err != nil {
			violations = append(
				violations,
				fieldViolation("email", err),
			)
		}
	}

	return violations
}
