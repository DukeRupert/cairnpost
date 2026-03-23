package service

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DealCreateInput struct {
	Title     string          `json:"title"`
	Stage     string          `json:"stage"`
	Value     decimal.Decimal `json:"value"`
	ContactID uuid.UUID       `json:"contact_id"`
	CompanyID *uuid.UUID      `json:"company_id"`
}

type DealUpdateInput struct {
	Title     *string          `json:"title"`
	Stage     *string          `json:"stage"`
	Value     *decimal.Decimal `json:"value"`
	ContactID *uuid.UUID       `json:"contact_id"`
	CompanyID *uuid.UUID       `json:"company_id"`
}

type DealStageInput struct {
	Stage string `json:"stage"`
}

type DealService interface {
	Create(ctx context.Context, orgID uuid.UUID, input DealCreateInput) (model.Deal, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Deal, error)
	List(ctx context.Context, orgID uuid.UUID, filter repository.DealFilter) ([]model.Deal, error)
	Update(ctx context.Context, orgID, id uuid.UUID, input DealUpdateInput) (model.Deal, error)
	UpdateStage(ctx context.Context, orgID, id uuid.UUID, input DealStageInput) (model.Deal, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type dealService struct {
	deals    repository.DealRepository
	contacts repository.ContactRepository
}

func NewDealService(deals repository.DealRepository, contacts repository.ContactRepository) DealService {
	return &dealService{deals: deals, contacts: contacts}
}

func (s *dealService) Create(ctx context.Context, orgID uuid.UUID, input DealCreateInput) (model.Deal, error) {
	if input.Title == "" {
		return model.Deal{}, &ValidationError{Field: "title", Message: "is required"}
	}

	if _, err := s.contacts.GetByID(ctx, orgID, input.ContactID); err != nil {
		return model.Deal{}, &ValidationError{Field: "contact_id", Message: "contact not found"}
	}

	stage := input.Stage
	if stage == "" {
		stage = model.DefaultStages[0]
	}

	d := model.Deal{
		OrgID:     orgID,
		Title:     input.Title,
		Stage:     stage,
		Value:     input.Value,
		ContactID: input.ContactID,
		CompanyID: input.CompanyID,
	}
	if err := s.deals.Create(ctx, &d); err != nil {
		return model.Deal{}, err
	}
	return d, nil
}

func (s *dealService) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Deal, error) {
	return s.deals.GetByID(ctx, orgID, id)
}

func (s *dealService) List(ctx context.Context, orgID uuid.UUID, filter repository.DealFilter) ([]model.Deal, error) {
	return s.deals.List(ctx, orgID, filter)
}

func (s *dealService) Update(ctx context.Context, orgID, id uuid.UUID, input DealUpdateInput) (model.Deal, error) {
	d, err := s.deals.GetByID(ctx, orgID, id)
	if err != nil {
		return model.Deal{}, err
	}

	if input.Title != nil {
		if *input.Title == "" {
			return model.Deal{}, &ValidationError{Field: "title", Message: "is required"}
		}
		d.Title = *input.Title
	}
	if input.Stage != nil {
		d.Stage = *input.Stage
	}
	if input.Value != nil {
		d.Value = *input.Value
	}
	if input.ContactID != nil {
		if _, err := s.contacts.GetByID(ctx, orgID, *input.ContactID); err != nil {
			return model.Deal{}, &ValidationError{Field: "contact_id", Message: "contact not found"}
		}
		d.ContactID = *input.ContactID
	}
	if input.CompanyID != nil {
		d.CompanyID = input.CompanyID
	}

	if err := s.deals.Update(ctx, &d); err != nil {
		return model.Deal{}, err
	}
	return d, nil
}

func (s *dealService) UpdateStage(ctx context.Context, orgID, id uuid.UUID, input DealStageInput) (model.Deal, error) {
	if input.Stage == "" {
		return model.Deal{}, &ValidationError{Field: "stage", Message: "is required"}
	}

	if err := s.deals.UpdateStage(ctx, orgID, id, input.Stage); err != nil {
		return model.Deal{}, err
	}
	return s.deals.GetByID(ctx, orgID, id)
}

func (s *dealService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.deals.Delete(ctx, orgID, id)
}
