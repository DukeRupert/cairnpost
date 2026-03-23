package service

import (
	"context"
	"time"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
)

type TaskCreateInput struct {
	Title      string     `json:"title"`
	DueDate    *time.Time `json:"due_date"`
	ContactID  *uuid.UUID `json:"contact_id"`
	DealID     *uuid.UUID `json:"deal_id"`
	AssignedTo uuid.UUID  `json:"assigned_to"`
}

type TaskUpdateInput struct {
	Title      *string    `json:"title"`
	DueDate    *time.Time `json:"due_date"`
	ContactID  *uuid.UUID `json:"contact_id"`
	DealID     *uuid.UUID `json:"deal_id"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
}

type TaskDoneInput struct {
	Done bool `json:"done"`
}

type TaskService interface {
	Create(ctx context.Context, orgID uuid.UUID, input TaskCreateInput) (model.Task, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Task, error)
	List(ctx context.Context, orgID uuid.UUID, filter repository.TaskFilter) ([]model.Task, error)
	Update(ctx context.Context, orgID, id uuid.UUID, input TaskUpdateInput) (model.Task, error)
	MarkDone(ctx context.Context, orgID, id uuid.UUID, input TaskDoneInput) (model.Task, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type taskService struct {
	tasks repository.TaskRepository
	users repository.UserRepository
}

func NewTaskService(tasks repository.TaskRepository, users repository.UserRepository) TaskService {
	return &taskService{tasks: tasks, users: users}
}

func (s *taskService) Create(ctx context.Context, orgID uuid.UUID, input TaskCreateInput) (model.Task, error) {
	if input.Title == "" {
		return model.Task{}, &ValidationError{Field: "title", Message: "is required"}
	}

	if _, err := s.users.GetByID(ctx, orgID, input.AssignedTo); err != nil {
		return model.Task{}, &ValidationError{Field: "assigned_to", Message: "user not found"}
	}

	t := model.Task{
		OrgID:      orgID,
		Title:      input.Title,
		DueDate:    input.DueDate,
		ContactID:  input.ContactID,
		DealID:     input.DealID,
		AssignedTo: input.AssignedTo,
	}
	if err := s.tasks.Create(ctx, &t); err != nil {
		return model.Task{}, err
	}
	return t, nil
}

func (s *taskService) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Task, error) {
	return s.tasks.GetByID(ctx, orgID, id)
}

func (s *taskService) List(ctx context.Context, orgID uuid.UUID, filter repository.TaskFilter) ([]model.Task, error) {
	return s.tasks.List(ctx, orgID, filter)
}

func (s *taskService) Update(ctx context.Context, orgID, id uuid.UUID, input TaskUpdateInput) (model.Task, error) {
	t, err := s.tasks.GetByID(ctx, orgID, id)
	if err != nil {
		return model.Task{}, err
	}

	if input.Title != nil {
		if *input.Title == "" {
			return model.Task{}, &ValidationError{Field: "title", Message: "is required"}
		}
		t.Title = *input.Title
	}
	if input.DueDate != nil {
		t.DueDate = input.DueDate
	}
	if input.ContactID != nil {
		t.ContactID = input.ContactID
	}
	if input.DealID != nil {
		t.DealID = input.DealID
	}
	if input.AssignedTo != nil {
		if _, err := s.users.GetByID(ctx, orgID, *input.AssignedTo); err != nil {
			return model.Task{}, &ValidationError{Field: "assigned_to", Message: "user not found"}
		}
		t.AssignedTo = *input.AssignedTo
	}

	if err := s.tasks.Update(ctx, &t); err != nil {
		return model.Task{}, err
	}
	return t, nil
}

func (s *taskService) MarkDone(ctx context.Context, orgID, id uuid.UUID, input TaskDoneInput) (model.Task, error) {
	if err := s.tasks.MarkDone(ctx, orgID, id, input.Done); err != nil {
		return model.Task{}, err
	}
	return s.tasks.GetByID(ctx, orgID, id)
}

func (s *taskService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.tasks.Delete(ctx, orgID, id)
}
