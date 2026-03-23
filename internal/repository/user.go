package repository

import (
	"context"
	"fmt"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserFilter struct {
	Role *model.Role
	Pagination
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, orgID uuid.UUID, email string) (model.User, error)
	List(ctx context.Context, orgID uuid.UUID, filter UserFilter) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	SetPasswordHash(ctx context.Context, orgID, userID uuid.UUID, hash string) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type pgUserRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &pgUserRepo{db: db}
}

func (r *pgUserRepo) Create(ctx context.Context, u *model.User) error {
	query := `INSERT INTO users (org_id, name, email, role, password_hash)
		VALUES (:org_id, :name, :email, :role, :password_hash)
		RETURNING id, created_at`
	rows, err := r.db.NamedQueryContext(ctx, query, u)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&u.ID, &u.CreatedAt)
	}
	return ErrNotFound
}

func (r *pgUserRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u,
		`SELECT * FROM users WHERE org_id = $1 AND id = $2`, orgID, id)
	return u, translateError(err)
}

func (r *pgUserRepo) GetByEmail(ctx context.Context, orgID uuid.UUID, email string) (model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u,
		`SELECT * FROM users WHERE org_id = $1 AND email = $2`, orgID, email)
	return u, translateError(err)
}

func (r *pgUserRepo) List(ctx context.Context, orgID uuid.UUID, f UserFilter) ([]model.User, error) {
	query := `SELECT * FROM users WHERE org_id = $1`
	args := []interface{}{orgID}
	n := 2

	if f.Role != nil {
		query += fmt.Sprintf(` AND role = $%d`, n)
		args = append(args, *f.Role)
		n++
	}

	query += ` ORDER BY name ASC`

	limit, offset := f.Pagination.normalize()
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	var users []model.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	return users, translateError(err)
}

func (r *pgUserRepo) Update(ctx context.Context, u *model.User) error {
	query := `UPDATE users SET name = :name, email = :email, role = :role
		WHERE org_id = :org_id AND id = :id`
	result, err := r.db.NamedExecContext(ctx, query, u)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgUserRepo) SetPasswordHash(ctx context.Context, orgID, userID uuid.UUID, hash string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE org_id = $2 AND id = $3`,
		hash, orgID, userID)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgUserRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM users WHERE org_id = $1 AND id = $2`, orgID, id)
	if err != nil {
		return translateError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
