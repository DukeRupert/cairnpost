package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TaskFilter struct {
	AssignedTo *uuid.UUID
	ContactID  *uuid.UUID
	DealID     *uuid.UUID
	Done       *bool
	DueBefore  *time.Time
	DueAfter   *time.Time
	Overdue    *bool // true = due_date < CURRENT_DATE AND NOT done
	Pagination
}

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Task, error)
	List(ctx context.Context, orgID uuid.UUID, filter TaskFilter) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	MarkDone(ctx context.Context, orgID, id uuid.UUID, done bool) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type pgTaskRepo struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) TaskRepository {
	return &pgTaskRepo{db: db}
}

func (r *pgTaskRepo) Create(ctx context.Context, t *model.Task) error {
	query := `INSERT INTO tasks (org_id, title, due_date, done, contact_id, deal_id, assigned_to)
		VALUES (:org_id, :title, :due_date, :done, :contact_id, :deal_id, :assigned_to)
		RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, t)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	}
	return ErrNotFound
}

func (r *pgTaskRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Task, error) {
	var t model.Task
	err := r.db.GetContext(ctx, &t,
		`SELECT * FROM tasks WHERE org_id = $1 AND id = $2`, orgID, id)
	return t, translateError(err)
}

func (r *pgTaskRepo) List(ctx context.Context, orgID uuid.UUID, f TaskFilter) ([]model.Task, error) {
	query := `SELECT * FROM tasks WHERE org_id = $1`
	args := []interface{}{orgID}
	n := 2

	if f.AssignedTo != nil {
		query += fmt.Sprintf(` AND assigned_to = $%d`, n)
		args = append(args, *f.AssignedTo)
		n++
	}
	if f.ContactID != nil {
		query += fmt.Sprintf(` AND contact_id = $%d`, n)
		args = append(args, *f.ContactID)
		n++
	}
	if f.DealID != nil {
		query += fmt.Sprintf(` AND deal_id = $%d`, n)
		args = append(args, *f.DealID)
		n++
	}
	if f.Done != nil {
		query += fmt.Sprintf(` AND done = $%d`, n)
		args = append(args, *f.Done)
		n++
	}
	if f.DueBefore != nil {
		query += fmt.Sprintf(` AND due_date <= $%d`, n)
		args = append(args, *f.DueBefore)
		n++
	}
	if f.DueAfter != nil {
		query += fmt.Sprintf(` AND due_date >= $%d`, n)
		args = append(args, *f.DueAfter)
		n++
	}
	if f.Overdue != nil && *f.Overdue {
		query += ` AND due_date < CURRENT_DATE AND NOT done`
	}

	query += ` ORDER BY done ASC, due_date ASC NULLS LAST`

	limit, offset := f.Pagination.normalize()
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	var tasks []model.Task
	err := r.db.SelectContext(ctx, &tasks, query, args...)
	return tasks, translateError(err)
}

func (r *pgTaskRepo) Update(ctx context.Context, t *model.Task) error {
	query := `UPDATE tasks SET
		title = :title, due_date = :due_date, done = :done,
		contact_id = :contact_id, deal_id = :deal_id,
		assigned_to = :assigned_to, updated_at = NOW()
		WHERE org_id = :org_id AND id = :id`
	result, err := r.db.NamedExecContext(ctx, query, t)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgTaskRepo) MarkDone(ctx context.Context, orgID, id uuid.UUID, done bool) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET done = $1, updated_at = NOW() WHERE org_id = $2 AND id = $3`,
		done, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgTaskRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE org_id = $1 AND id = $2`, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
