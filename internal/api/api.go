package api

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userPkg "gitlab.ozon.dev/iTukaev/homework/internal/pkg/core/user"
	localPkg "gitlab.ozon.dev/iTukaev/homework/internal/pkg/core/user/cache/local"
	"gitlab.ozon.dev/iTukaev/homework/internal/pkg/core/user/models"
	pb "gitlab.ozon.dev/iTukaev/homework/pkg/api"
)

const (
	contextTimeout = 5 * time.Second
)

func New(user userPkg.Interface) pb.UserServer {
	return &implementation{
		user: user,
	}
}

type implementation struct {
	user userPkg.Interface
	pb.UnimplementedUserServer
}

func (i *implementation) UserCreate(ctx context.Context, in *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	if err := i.user.Create(ctxWithTimeout, models.User{
		Name:     in.GetName(),
		Password: in.GetPassword(),
	}); err != nil {
		log.Printf("user [%s] create: %v", in.GetName(), err)

		switch {
		case errors.Is(err, userPkg.ErrValidation):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, localPkg.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, localPkg.ErrTimeout):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UserCreateResponse{}, nil
}

func (i *implementation) UserUpdate(ctx context.Context, in *pb.UserUpdateRequest) (*pb.UserUpdateResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	if err := i.user.Update(ctxWithTimeout, models.User{
		Name:     in.GetName(),
		Password: in.GetPassword(),
	}); err != nil {
		log.Printf("user [%s] update: %v", in.GetName(), err)

		switch {
		case errors.Is(err, userPkg.ErrValidation), errors.Is(err, localPkg.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, localPkg.ErrTimeout):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UserUpdateResponse{}, nil
}

func (i *implementation) UserDelete(ctx context.Context, in *pb.UserDeleteRequest) (*pb.UserDeleteResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	if err := i.user.Delete(ctxWithTimeout, in.GetName()); err != nil {
		log.Printf("user [%s] delete: %v", in.GetName(), err)

		switch {
		case errors.Is(err, userPkg.ErrValidation), errors.Is(err, localPkg.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, localPkg.ErrTimeout):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UserDeleteResponse{}, nil
}

func (i *implementation) UserGet(ctx context.Context, in *pb.UserGetRequest) (*pb.UserGetResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	user, err := i.user.Get(ctxWithTimeout, in.GetName())
	if err != nil {
		log.Printf("user [%s] get: %v", in.GetName(), err)

		switch {
		case errors.Is(err, userPkg.ErrValidation), errors.Is(err, localPkg.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, localPkg.ErrTimeout):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UserGetResponse{
		User: &pb.UserGetResponse_User{
			Name:     user.Name,
			Password: user.Password,
		},
	}, nil
}

func (i *implementation) UserList(ctx context.Context, _ *pb.UserListRequest) (*pb.UserListResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	users, err := i.user.List(ctxWithTimeout)
	if errors.Is(err, localPkg.ErrTimeout) {
		return &pb.UserListResponse{}, status.Error(codes.DeadlineExceeded, err.Error())
	}

	resp := make([]*pb.UserListResponse_User, 0, len(users))
	for _, user := range users {
		resp = append(resp, &pb.UserListResponse_User{
			Name:     user.Name,
			Password: user.Password,
		})
	}
	return &pb.UserListResponse{
		Users: resp,
	}, nil
}
