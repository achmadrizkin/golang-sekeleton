package endpoint

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/internal/usecase"
)

// ---- request / response ----

type AddUserRequest struct{ User *domain.User }
type AddUserResponse struct{ Err error }

type GetUserByIDRequest struct{ ID string }
type GetUserByIDResponse struct {
	User *domain.User
	Err  error
}

type GetAllUsersRequest struct{}
type GetAllUsersResponse struct {
	Users []*domain.User
	Err   error
}

type GetAllUsersPaginatedRequest struct {
	Page     int
	PageSize int
}
type GetAllUsersPaginatedResponse struct {
	PaginatedUsers *domain.PaginatedUsers
	Err            error
}

type UpdateUserRequest struct{ User *domain.User }
type UpdateUserResponse struct{ Err error }

type DeleteUserRequest struct{ ID string }
type DeleteUserResponse struct{ Err error }

type SearchUsersRequest struct{ Query string }
type SearchUsersResponse struct {
	Users []*domain.User
	Err   error
}

type ProcessUserCreatedRequest struct{ ID string }
type ProcessUserCreatedResponse struct{ Err error }

type ProcessUserDeletedRequest struct{ ID string }
type ProcessUserDeletedResponse struct{ Err error }

// ---- factories ----

func MakeAddUserEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(AddUserRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		err := uc.Add(ctx, req.User)
		return AddUserResponse{Err: err}, err
	}
}

func MakeGetUserByIDEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetUserByIDRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		user, err := uc.GetByID(ctx, req.ID)
		return GetUserByIDResponse{User: user, Err: err}, err
	}
}

func MakeGetAllUsersEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		users, err := uc.GetAll(ctx)
		return GetAllUsersResponse{Users: users, Err: err}, err
	}
}

func MakeGetAllUsersPaginatedEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetAllUsersPaginatedRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		result, err := uc.GetAllPaginated(ctx, &domain.PaginationRequest{Page: req.Page, PageSize: req.PageSize})
		return GetAllUsersPaginatedResponse{PaginatedUsers: result, Err: err}, err
	}
}

func MakeUpdateUserEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(UpdateUserRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		err := uc.Update(ctx, req.User)
		return UpdateUserResponse{Err: err}, err
	}
}

func MakeDeleteUserEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(DeleteUserRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		err := uc.Delete(ctx, req.ID)
		return DeleteUserResponse{Err: err}, err
	}
}

func MakeSearchUsersEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(SearchUsersRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		users, err := uc.Search(ctx, req.Query)
		return SearchUsersResponse{Users: users, Err: err}, err
	}
}

func MakeProcessUserCreatedEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(ProcessUserCreatedRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		err := uc.ProcessUserCreated(ctx, req.ID)
		return ProcessUserCreatedResponse{Err: err}, err
	}
}

func MakeProcessUserDeletedEndpoint(uc usecase.UserUseCaseInterface) Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(ProcessUserDeletedRequest)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		err := uc.ProcessUserDeleted(ctx, req.ID)
		return ProcessUserDeletedResponse{Err: err}, err
	}
}

// ---- singleflight keys (reads + event processing only) ----

func GetUserByIDKey(request interface{}) string {
	req, ok := request.(GetUserByIDRequest)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("get_user_by_id:%s", req.ID)
}

func GetAllUsersKey(_ interface{}) string { return "get_all_users" }

func SearchUsersKey(request interface{}) string {
	req, ok := request.(SearchUsersRequest)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("search_users:%s", req.Query)
}

func ProcessUserCreatedKey(request interface{}) string {
	req, ok := request.(ProcessUserCreatedRequest)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("process_user_created:%s", req.ID)
}

func ProcessUserDeletedKey(request interface{}) string {
	req, ok := request.(ProcessUserDeletedRequest)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("process_user_deleted:%s", req.ID)
}

// ---- bundle ----

// UserEndpoints groups every endpoint used by delivery/grpc, delivery/http,
// and delivery/messaging.
type UserEndpoints struct {
	AddUserEndpoint              Endpoint
	GetUserByIDEndpoint          Endpoint
	GetAllUsersEndpoint          Endpoint
	GetAllUsersPaginatedEndpoint Endpoint
	UpdateUserEndpoint           Endpoint
	DeleteUserEndpoint           Endpoint
	SearchUsersEndpoint          Endpoint
	ProcessUserCreatedEndpoint   Endpoint
	ProcessUserDeletedEndpoint   Endpoint
}

// MakeUserEndpoints wires every operation, wrapping reads and event
// processing in singleflight.
func MakeUserEndpoints(uc usecase.UserUseCaseInterface) UserEndpoints {
	return UserEndpoints{
		AddUserEndpoint:              MakeAddUserEndpoint(uc),
		GetUserByIDEndpoint:          withSingleflight(MakeGetUserByIDEndpoint(uc), GetUserByIDKey),
		GetAllUsersEndpoint:          withSingleflight(MakeGetAllUsersEndpoint(uc), GetAllUsersKey),
		GetAllUsersPaginatedEndpoint: MakeGetAllUsersPaginatedEndpoint(uc),
		UpdateUserEndpoint:           MakeUpdateUserEndpoint(uc),
		DeleteUserEndpoint:           MakeDeleteUserEndpoint(uc),
		SearchUsersEndpoint:          withSingleflight(MakeSearchUsersEndpoint(uc), SearchUsersKey),
		ProcessUserCreatedEndpoint:   withSingleflight(MakeProcessUserCreatedEndpoint(uc), ProcessUserCreatedKey),
		ProcessUserDeletedEndpoint:   withSingleflight(MakeProcessUserDeletedEndpoint(uc), ProcessUserDeletedKey),
	}
}
