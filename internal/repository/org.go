package repository

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrgRepository interface {
	Create(ctx context.Context, org *model.Org) error
	GetByID(ctx context.Context, id uuid.UUID) (model.Org, error)
	GetBySlug(ctx context.Context, slug string) (model.Org, error)
	Update(ctx context.Context, org *model.Org) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgOrgRepo struct {
	db *sqlx.DB
}

func NewOrgRepository(db *sqlx.DB) OrgRepository {
	return &pgOrgRepo{db: db}
}

func (r *pgOrgRepo) Create(ctx context.Context, org *model.Org) error {
	query := `INSERT INTO orgs (name, slug)
		VALUES (:name, :slug)
		RETURNING id, created_at`
	rows, err := r.db.NamedQueryContext(ctx, query, org)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&org.ID, &org.CreatedAt)
	}
	return ErrNotFound
}

func (r *pgOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (model.Org, error) {
	var org model.Org
	err := r.db.GetContext(ctx, &org, `SELECT * FROM orgs WHERE id = $1`, id)
	return org, translateError(err)
}

func (r *pgOrgRepo) GetBySlug(ctx context.Context, slug string) (model.Org, error) {
	var org model.Org
	err := r.db.GetContext(ctx, &org, `SELECT * FROM orgs WHERE slug = $1`, slug)
	return org, translateError(err)
}

func (r *pgOrgRepo) Update(ctx context.Context, org *model.Org) error {
	query := `UPDATE orgs SET name = :name, slug = :slug WHERE id = :id`
	result, err := r.db.NamedExecContext(ctx, query, org)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgOrgRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM orgs WHERE id = $1`, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
