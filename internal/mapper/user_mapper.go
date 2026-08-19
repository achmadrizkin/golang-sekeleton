// Package mapper isolates protobuf wrapper types (StringValue, BoolValue,
// Timestamp) to the boundary between delivery/grpc and the rest of the
// service — domain.User never sees a wrapperspb type.
package mapper

import (
	"github.com/fauzie/golang-sekeleton/internal/domain"
	pb "github.com/fauzie/golang-sekeleton/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// UserMapper converts between domain.User and pb.UserModel.
type UserMapper struct{}

// NewUserMapper builds a UserMapper.
func NewUserMapper() *UserMapper { return &UserMapper{} }

// ProtoToDomain converts proto UserModel to domain User.
func (m *UserMapper) ProtoToDomain(proto *pb.UserModel) *domain.User {
	user := &domain.User{}

	if proto.GetId() != nil {
		user.ID = proto.GetId().GetValue()
	}
	if proto.GetUsername() != nil {
		user.Username = proto.GetUsername().GetValue()
	}
	if proto.GetEmail() != nil {
		user.Email = proto.GetEmail().GetValue()
	}
	if proto.GetPassword() != nil {
		user.Password = proto.GetPassword().GetValue()
	}
	if proto.GetFullName() != nil {
		user.FullName = proto.GetFullName().GetValue()
	}
	if proto.GetAvatar() != nil {
		avatar := proto.GetAvatar().GetValue()
		user.Avatar = &avatar
	}
	if proto.GetRole() != nil {
		role := proto.GetRole().GetValue()
		user.Role = &role
	}
	if proto.GetIsActive() != nil {
		user.IsActive = proto.GetIsActive().GetValue()
	}
	if proto.GetEmailVerified() != nil {
		user.EmailVerified = proto.GetEmailVerified().GetValue()
	}

	return user
}

// DomainToProto converts domain User to proto UserModel. Password is only
// populated when non-empty, so a hash never rides along on a read response.
func (m *UserMapper) DomainToProto(user *domain.User) *pb.UserModel {
	proto := &pb.UserModel{
		Id:            wrapperspb.String(user.ID),
		Username:      wrapperspb.String(user.Username),
		Email:         wrapperspb.String(user.Email),
		FullName:      wrapperspb.String(user.FullName),
		IsActive:      wrapperspb.Bool(user.IsActive),
		EmailVerified: wrapperspb.Bool(user.EmailVerified),
	}

	if user.Password != "" {
		proto.Password = wrapperspb.String(user.Password)
	}
	if user.Role != nil {
		proto.Role = wrapperspb.String(*user.Role)
	}
	if user.Avatar != nil {
		proto.Avatar = wrapperspb.String(*user.Avatar)
	}
	if !user.CreatedAt.IsZero() {
		proto.CreatedAt = timestamppb.New(user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		proto.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}

	return proto
}

// DomainListToProto converts a slice of domain Users to proto UserModels.
func (m *UserMapper) DomainListToProto(users []*domain.User) []*pb.UserModel {
	out := make([]*pb.UserModel, 0, len(users))
	for _, u := range users {
		out = append(out, m.DomainToProto(u))
	}
	return out
}
