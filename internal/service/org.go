package service

import (
	"context"
	"regexp"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type OrgCreateInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type OrgUpdateInput struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

type OrgService interface {
	Create(ctx context.Context, input OrgCreateInput) (model.Org, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.Org, error)
	GetBySlug(ctx context.Context, slug string) (model.Org, error)
	Update(ctx context.Context, id uuid.UUID, input OrgUpdateInput) (model.Org, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type orgService struct {
	orgs repository.OrgRepository
}

func NewOrgService(orgs repository.OrgRepository) OrgService {
	return &orgService{orgs: orgs}
}

func (s *orgService) Create(ctx context.Context, input OrgCreateInput) (model.Org, error) {
	if input.Name == "" {
		return model.Org{}, &ValidationError{Field: "name", Message: "is required"}
	}
	if !slugPattern.MatchString(input.Slug) {
		return model.Org{}, &ValidationError{Field: "slug", Message: "must be lowercase alphanumeric with hyphens"}
	}

	org := model.Org{
		Name: input.Name,
		Slug: input.Slug,
	}
	if err := s.orgs.Create(ctx, &org); err != nil {
		return model.Org{}, err
	}
	return org, nil
}

func (s *orgService) GetByID(ctx context.Context, id uuid.UUID) (model.Org, error) {
	return s.orgs.GetByID(ctx, id)
}

func (s *orgService) GetBySlug(ctx context.Context, slug string) (model.Org, error) {
	return s.orgs.GetBySlug(ctx, slug)
}

func (s *orgService) Update(ctx context.Context, id uuid.UUID, input OrgUpdateInput) (model.Org, error) {
	org, err := s.orgs.GetByID(ctx, id)
	if err != nil {
		return model.Org{}, err
	}

	if input.Name != nil {
		if *input.Name == "" {
			return model.Org{}, &ValidationError{Field: "name", Message: "is required"}
		}
		org.Name = *input.Name
	}
	if input.Slug != nil {
		if !slugPattern.MatchString(*input.Slug) {
			return model.Org{}, &ValidationError{Field: "slug", Message: "must be lowercase alphanumeric with hyphens"}
		}
		org.Slug = *input.Slug
	}

	if err := s.orgs.Update(ctx, &org); err != nil {
		return model.Org{}, err
	}
	return org, nil
}

func (s *orgService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.orgs.Delete(ctx, id)
}
