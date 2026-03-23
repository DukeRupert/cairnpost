package service

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ContactCreateInput struct {
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Tags      []string   `json:"tags"`
	CompanyID *uuid.UUID `json:"company_id"`
}

type ContactUpdateInput struct {
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	Email     *string    `json:"email"`
	Phone     *string    `json:"phone"`
	Tags      *[]string  `json:"tags"`
	CompanyID *uuid.UUID `json:"company_id"`
}

type ContactService interface {
	Create(ctx context.Context, orgID uuid.UUID, input ContactCreateInput) (model.Contact, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Contact, error)
	List(ctx context.Context, orgID uuid.UUID, filter repository.ContactFilter) ([]model.Contact, error)
	Update(ctx context.Context, orgID, id uuid.UUID, input ContactUpdateInput) (model.Contact, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type contactService struct {
	contacts  repository.ContactRepository
	companies repository.CompanyRepository
}

func NewContactService(contacts repository.ContactRepository, companies repository.CompanyRepository) ContactService {
	return &contactService{contacts: contacts, companies: companies}
}

func (s *contactService) Create(ctx context.Context, orgID uuid.UUID, input ContactCreateInput) (model.Contact, error) {
	if input.FirstName == "" {
		return model.Contact{}, &ValidationError{Field: "first_name", Message: "is required"}
	}

	if input.CompanyID != nil {
		if _, err := s.companies.GetByID(ctx, orgID, *input.CompanyID); err != nil {
			return model.Contact{}, &ValidationError{Field: "company_id", Message: "company not found"}
		}
	}

	tags := pq.StringArray(input.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}

	c := model.Contact{
		OrgID:     orgID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Tags:      tags,
		CompanyID: input.CompanyID,
	}
	if err := s.contacts.Create(ctx, &c); err != nil {
		return model.Contact{}, err
	}
	return c, nil
}

func (s *contactService) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Contact, error) {
	return s.contacts.GetByID(ctx, orgID, id)
}

func (s *contactService) List(ctx context.Context, orgID uuid.UUID, filter repository.ContactFilter) ([]model.Contact, error) {
	return s.contacts.List(ctx, orgID, filter)
}

func (s *contactService) Update(ctx context.Context, orgID, id uuid.UUID, input ContactUpdateInput) (model.Contact, error) {
	c, err := s.contacts.GetByID(ctx, orgID, id)
	if err != nil {
		return model.Contact{}, err
	}

	if input.FirstName != nil {
		if *input.FirstName == "" {
			return model.Contact{}, &ValidationError{Field: "first_name", Message: "is required"}
		}
		c.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		c.LastName = *input.LastName
	}
	if input.Email != nil {
		c.Email = *input.Email
	}
	if input.Phone != nil {
		c.Phone = *input.Phone
	}
	if input.Tags != nil {
		c.Tags = pq.StringArray(*input.Tags)
	}
	if input.CompanyID != nil {
		if _, err := s.companies.GetByID(ctx, orgID, *input.CompanyID); err != nil {
			return model.Contact{}, &ValidationError{Field: "company_id", Message: "company not found"}
		}
		c.CompanyID = input.CompanyID
	}

	if err := s.contacts.Update(ctx, &c); err != nil {
		return model.Contact{}, err
	}
	return c, nil
}

func (s *contactService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.contacts.Delete(ctx, orgID, id)
}
