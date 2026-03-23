package service

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
)

type UserCreateInput struct {
	Name  string     `json:"name"`
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}

type UserUpdateInput struct {
	Name  *string     `json:"name"`
	Email *string     `json:"email"`
	Role  *model.Role `json:"role"`
}

type UserService interface {
	Create(ctx context.Context, orgID uuid.UUID, input UserCreateInput) (model.User, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.User, error)
	List(ctx context.Context, orgID uuid.UUID, filter repository.UserFilter) ([]model.User, error)
	Update(ctx context.Context, orgID, id uuid.UUID, input UserUpdateInput) (model.User, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) Create(ctx context.Context, orgID uuid.UUID, input UserCreateInput) (model.User, error) {
	if input.Name == "" {
		return model.User{}, &ValidationError{Field: "name", Message: "is required"}
	}
	if input.Email == "" {
		return model.User{}, &ValidationError{Field: "email", Message: "is required"}
	}
	if err := validateRole(input.Role); err != nil {
		return model.User{}, err
	}

	u := model.User{
		OrgID: orgID,
		Name:  input.Name,
		Email: input.Email,
		Role:  input.Role,
	}
	if err := s.users.Create(ctx, &u); err != nil {
		return model.User{}, err
	}
	return u, nil
}

func (s *userService) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.User, error) {
	return s.users.GetByID(ctx, orgID, id)
}

func (s *userService) List(ctx context.Context, orgID uuid.UUID, filter repository.UserFilter) ([]model.User, error) {
	return s.users.List(ctx, orgID, filter)
}

func (s *userService) Update(ctx context.Context, orgID, id uuid.UUID, input UserUpdateInput) (model.User, error) {
	u, err := s.users.GetByID(ctx, orgID, id)
	if err != nil {
		return model.User{}, err
	}

	if input.Name != nil {
		if *input.Name == "" {
			return model.User{}, &ValidationError{Field: "name", Message: "is required"}
		}
		u.Name = *input.Name
	}
	if input.Email != nil {
		if *input.Email == "" {
			return model.User{}, &ValidationError{Field: "email", Message: "is required"}
		}
		u.Email = *input.Email
	}
	if input.Role != nil {
		if err := validateRole(*input.Role); err != nil {
			return model.User{}, err
		}
		u.Role = *input.Role
	}

	if err := s.users.Update(ctx, &u); err != nil {
		return model.User{}, err
	}
	return u, nil
}

func (s *userService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.users.Delete(ctx, orgID, id)
}

func validateRole(role model.Role) error {
	if role != model.RoleAdmin && role != model.RoleMember {
		return &ValidationError{Field: "role", Message: "must be admin or member"}
	}
	return nil
}
