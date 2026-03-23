package repository

import (
	"context"
	"fmt"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ContactFilter struct {
	Search    *string
	Tag       *string
	CompanyID *uuid.UUID
	Pagination
}

type ContactRepository interface {
	Create(ctx context.Context, contact *model.Contact) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Contact, error)
	List(ctx context.Context, orgID uuid.UUID, filter ContactFilter) ([]model.Contact, error)
	Update(ctx context.Context, contact *model.Contact) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type pgContactRepo struct {
	db *sqlx.DB
}

func NewContactRepository(db *sqlx.DB) ContactRepository {
	return &pgContactRepo{db: db}
}

func (r *pgContactRepo) Create(ctx context.Context, c *model.Contact) error {
	query := `INSERT INTO contacts (org_id, first_name, last_name, email, phone, tags, company_id)
		VALUES (:org_id, :first_name, :last_name, :email, :phone, :tags, :company_id)
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

func (r *pgContactRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.Contact, error) {
	var c model.Contact
	err := r.db.GetContext(ctx, &c,
		`SELECT * FROM contacts WHERE org_id = $1 AND id = $2`, orgID, id)
	return c, translateError(err)
}

func (r *pgContactRepo) List(ctx context.Context, orgID uuid.UUID, f ContactFilter) ([]model.Contact, error) {
	query := `SELECT * FROM contacts WHERE org_id = $1`
	args := []interface{}{orgID}
	n := 2

	if f.Search != nil {
		query += fmt.Sprintf(` AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)`, n, n, n)
		args = append(args, "%"+*f.Search+"%")
		n++
	}
	if f.Tag != nil {
		query += fmt.Sprintf(` AND $%d = ANY(tags)`, n)
		args = append(args, *f.Tag)
		n++
	}
	if f.CompanyID != nil {
		query += fmt.Sprintf(` AND company_id = $%d`, n)
		args = append(args, *f.CompanyID)
		n++
	}

	query += ` ORDER BY last_name ASC, first_name ASC`

	limit, offset := f.Pagination.normalize()
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	var contacts []model.Contact
	err := r.db.SelectContext(ctx, &contacts, query, args...)
	return contacts, translateError(err)
}

func (r *pgContactRepo) Update(ctx context.Context, c *model.Contact) error {
	query := `UPDATE contacts SET
		first_name = :first_name, last_name = :last_name,
		email = :email, phone = :phone, tags = :tags,
		company_id = :company_id, updated_at = NOW()
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

func (r *pgContactRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM contacts WHERE org_id = $1 AND id = $2`, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
