package repository

import (
	"context"
	"fmt"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DealFilter struct {
	Stage     *string
	ContactID *uuid.UUID
	CompanyID *uuid.UUID
	Open      *bool // true = closed_at IS NULL, false = closed_at IS NOT NULL
	Pagination
}

type DealRepository interface {
	Create(ctx context.Context, deal *model.Deal) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Deal, error)
	List(ctx context.Context, orgID uuid.UUID, filter DealFilter) ([]model.Deal, error)
	Update(ctx context.Context, deal *model.Deal) error
	UpdateStage(ctx context.Context, orgID, id uuid.UUID, stage string) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type pgDealRepo struct {
	db *sqlx.DB
}

func NewDealRepository(db *sqlx.DB) DealRepository {
	return &pgDealRepo{db: db}
}

func (r *pgDealRepo) Create(ctx context.Context, d *model.Deal) error {
	query := `INSERT INTO deals (org_id, title, stage, value, contact_id, company_id)
		VALUES (:org_id, :title, :stage, :value, :contact_id, :company_id)
		RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, d)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	}
	return ErrNotFound
}

func (r *pgDealRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Deal, error) {
	var d model.Deal
	err := r.db.GetContext(ctx, &d,
		`SELECT * FROM deals WHERE org_id = $1 AND id = $2`, orgID, id)
	return d, translateError(err)
}

func (r *pgDealRepo) List(ctx context.Context, orgID uuid.UUID, f DealFilter) ([]model.Deal, error) {
	query := `SELECT * FROM deals WHERE org_id = $1`
	args := []interface{}{orgID}
	n := 2

	if f.Stage != nil {
		query += fmt.Sprintf(` AND stage = $%d`, n)
		args = append(args, *f.Stage)
		n++
	}
	if f.ContactID != nil {
		query += fmt.Sprintf(` AND contact_id = $%d`, n)
		args = append(args, *f.ContactID)
		n++
	}
	if f.CompanyID != nil {
		query += fmt.Sprintf(` AND company_id = $%d`, n)
		args = append(args, *f.CompanyID)
		n++
	}
	if f.Open != nil {
		if *f.Open {
			query += ` AND closed_at IS NULL`
		} else {
			query += ` AND closed_at IS NOT NULL`
		}
	}

	query += ` ORDER BY created_at DESC`

	limit, offset := f.Pagination.normalize()
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	var deals []model.Deal
	err := r.db.SelectContext(ctx, &deals, query, args...)
	return deals, translateError(err)
}

func (r *pgDealRepo) Update(ctx context.Context, d *model.Deal) error {
	query := `UPDATE deals SET
		title = :title, stage = :stage, value = :value,
		contact_id = :contact_id, company_id = :company_id,
		closed_at = :closed_at, updated_at = NOW()
		WHERE org_id = :org_id AND id = :id`
	result, err := r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgDealRepo) UpdateStage(ctx context.Context, orgID, id uuid.UUID, stage string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE deals SET stage = $1, updated_at = NOW() WHERE org_id = $2 AND id = $3`,
		stage, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgDealRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM deals WHERE org_id = $1 AND id = $2`, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
