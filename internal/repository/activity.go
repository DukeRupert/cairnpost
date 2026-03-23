package repository

import (
	"context"
	"fmt"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ActivityFilter struct {
	ContactID *uuid.UUID
	DealID    *uuid.UUID
	Type      *model.ActivityType
	Pagination
}

type ActivityRepository interface {
	Create(ctx context.Context, activity *model.Activity) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Activity, error)
	List(ctx context.Context, orgID uuid.UUID, filter ActivityFilter) ([]model.Activity, error)
}

type pgActivityRepo struct {
	db *sqlx.DB
}

func NewActivityRepository(db *sqlx.DB) ActivityRepository {
	return &pgActivityRepo{db: db}
}

func (r *pgActivityRepo) Create(ctx context.Context, a *model.Activity) error {
	query := `INSERT INTO activities (org_id, type, body, contact_id, deal_id, user_id, occurred_at)
		VALUES (:org_id, :type, :body, :contact_id, :deal_id, :user_id, :occurred_at)
		RETURNING id, created_at`
	rows, err := r.db.NamedQueryContext(ctx, query, a)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&a.ID, &a.CreatedAt)
	}
	return ErrNotFound
}

func (r *pgActivityRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Activity, error) {
	var a model.Activity
	err := r.db.GetContext(ctx, &a,
		`SELECT * FROM activities WHERE org_id = $1 AND id = $2`, orgID, id)
	return a, translateError(err)
}

func (r *pgActivityRepo) List(ctx context.Context, orgID uuid.UUID, f ActivityFilter) ([]model.Activity, error) {
	query := `SELECT * FROM activities WHERE org_id = $1`
	args := []interface{}{orgID}
	n := 2

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
	if f.Type != nil {
		query += fmt.Sprintf(` AND type = $%d`, n)
		args = append(args, *f.Type)
		n++
	}

	query += ` ORDER BY occurred_at DESC`

	limit, offset := f.Pagination.normalize()
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	var activities []model.Activity
	err := r.db.SelectContext(ctx, &activities, query, args...)
	return activities, translateError(err)
}
