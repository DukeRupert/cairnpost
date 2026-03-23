package repository

import (
	"context"
	"fmt"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CompanyFilter struct {
	Search *string
	Pagination
}

type CompanyRepository interface {
	Create(ctx context.Context, company *model.Company) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Company, error)
	List(ctx context.Context, orgID uuid.UUID, filter CompanyFilter) ([]model.Company, error)
	Update(ctx context.Context, company *model.Company) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type pgCompanyRepo struct {
	db *sqlx.DB
}

func NewCompanyRepository(db *sqlx.DB) CompanyRepository {
	return &pgCompanyRepo{db: db}
}

func (r *pgCompanyRepo) Create(ctx context.Context, c *model.Company) error {
	query := `INSERT INTO companies (org_id, name, address, website, notes)
		VALUES (:org_id, :name, :address, :website, :notes)
		RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, c)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	}
	return ErrNotFound
}

func (r *pgCompanyRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Company, error) {
	var c model.Company
	err := r.db.GetContext(ctx, &c,
		`SELECT * FROM companies WHERE org_id = $1 AND id = $2`, orgID, id)
	return c, translateError(err)
}

func (r *pgCompanyRepo) List(ctx context.Context, orgID uuid.UUID, f CompanyFilter) ([]model.Company, error) {
	query := `SELECT * FROM companies WHERE org_id = $1`
	args := []interface{}{orgID}
	n := 2

	if f.Search != nil {
		query += fmt.Sprintf(` AND name ILIKE $%d`, n)
		args = append(args, "%"+*f.Search+"%")
		n++
	}

	query += ` ORDER BY name ASC`

	limit, offset := f.Pagination.normalize()
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	var companies []model.Company
	err := r.db.SelectContext(ctx, &companies, query, args...)
	return companies, translateError(err)
}

func (r *pgCompanyRepo) Update(ctx context.Context, c *model.Company) error {
	query := `UPDATE companies SET
		name = :name, address = :address, website = :website,
		notes = :notes, updated_at = NOW()
		WHERE org_id = :org_id AND id = :id`
	result, err := r.db.NamedExecContext(ctx, query, c)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgCompanyRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM companies WHERE org_id = $1 AND id = $2`, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
