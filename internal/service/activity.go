package service

import (
	"context"
	"time"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
)

type ActivityCreateInput struct {
	Type       model.ActivityType `json:"type"`
	Body       string             `json:"body"`
	ContactID  uuid.UUID          `json:"contact_id"`
	DealID     *uuid.UUID         `json:"deal_id"`
	UserID     uuid.UUID          `json:"user_id"`
	OccurredAt *time.Time         `json:"occurred_at"`
}

type ActivityService interface {
	Create(ctx context.Context, orgID uuid.UUID, input ActivityCreateInput) (model.Activity, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Activity, error)
	List(ctx context.Context, orgID uuid.UUID, filter repository.ActivityFilter) ([]model.Activity, error)
}

type activityService struct {
	activities repository.ActivityRepository
	contacts   repository.ContactRepository
}

func NewActivityService(activities repository.ActivityRepository, contacts repository.ContactRepository) ActivityService {
	return &activityService{activities: activities, contacts: contacts}
}

func (s *activityService) Create(ctx context.Context, orgID uuid.UUID, input ActivityCreateInput) (model.Activity, error) {
	if err := validateActivityType(input.Type); err != nil {
		return model.Activity{}, err
	}
	if input.Body == "" {
		return model.Activity{}, &ValidationError{Field: "body", Message: "is required"}
	}

	if _, err := s.contacts.GetByID(ctx, orgID, input.ContactID); err != nil {
		return model.Activity{}, &ValidationError{Field: "contact_id", Message: "contact not found"}
	}

	occurredAt := time.Now()
	if input.OccurredAt != nil {
		occurredAt = *input.OccurredAt
	}

	a := model.Activity{
		OrgID:      orgID,
		Type:       input.Type,
		Body:       input.Body,
		ContactID:  input.ContactID,
		DealID:     input.DealID,
		UserID:     input.UserID,
		OccurredAt: occurredAt,
	}
	if err := s.activities.Create(ctx, &a); err != nil {
		return model.Activity{}, err
	}
	return a, nil
}

func (s *activityService) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Activity, error) {
	return s.activities.GetByID(ctx, orgID, id)
}

func (s *activityService) List(ctx context.Context, orgID uuid.UUID, filter repository.ActivityFilter) ([]model.Activity, error) {
	return s.activities.List(ctx, orgID, filter)
}

func validateActivityType(t model.ActivityType) error {
	switch t {
	case model.ActivityNote, model.ActivityCall, model.ActivityEmail,
		model.ActivitySMS, model.ActivitySiteVisit:
		return nil
	}
	return &ValidationError{Field: "type", Message: "must be note, call, email, sms, or site_visit"}
}
