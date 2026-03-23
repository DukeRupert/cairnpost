package service

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
)

type CompanyCreateInput struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Website string `json:"website"`
	Notes   string `json:"notes"`
}

type CompanyUpdateInput struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
	Website *string `json:"website"`
	Notes   *string `json:"notes"`
}

type CompanyService interface {
	Create(ctx context.Context, orgID uuid.UUID, input CompanyCreateInput) (model.Company, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Company, error)
	List(ctx context.Context, orgID uuid.UUID, filter repository.CompanyFilter) ([]model.Company, error)
	Update(ctx context.Context, orgID, id uuid.UUID, input CompanyUpdateInput) (model.Company, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type companyService struct {
	companies repository.CompanyRepository
}

func NewCompanyService(companies repository.CompanyRepository) CompanyService {
	return &companyService{companies: companies}
}

func (s *companyService) Create(ctx context.Context, orgID uuid.UUID, input CompanyCreateInput) (model.Company, error) {
	if input.Name == "" {
		return model.Company{}, &ValidationError{Field: "name", Message: "is required"}
	}

	c := model.Company{
		OrgID:   orgID,
		Name:    input.Name,
		Address: input.Address,
		Website: input.Website,
		Notes:   input.Notes,
	}
	if err := s.companies.Create(ctx, &c); err != nil {
		return model.Company{}, err
	}
	return c, nil
}

func (s *companyService) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Company, error) {
	return s.companies.GetByID(ctx, orgID, id)
}

func (s *companyService) List(ctx context.Context, orgID uuid.UUID, filter repository.CompanyFilter) ([]model.Company, error) {
	return s.companies.List(ctx, orgID, filter)
}

func (s *companyService) Update(ctx context.Context, orgID, id uuid.UUID, input CompanyUpdateInput) (model.Company, error) {
	c, err := s.companies.GetByID(ctx, orgID, id)
	if err != nil {
		return model.Company{}, err
	}

	if input.Name != nil {
		if *input.Name == "" {
			return model.Company{}, &ValidationError{Field: "name", Message: "is required"}
		}
		c.Name = *input.Name
	}
	if input.Address != nil {
		c.Address = *input.Address
	}
	if input.Website != nil {
		c.Website = *input.Website
	}
	if input.Notes != nil {
		c.Notes = *input.Notes
	}

	if err := s.companies.Update(ctx, &c); err != nil {
		return model.Company{}, err
	}
	return c, nil
}

func (s *companyService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.companies.Delete(ctx, orgID, id)
}
