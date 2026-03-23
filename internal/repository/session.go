package repository

import (
	"context"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (model.Session, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type pgSessionRepo struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &pgSessionRepo{db: db}
}

func (r *pgSessionRepo) Create(ctx context.Context, s *model.Session) error {
	query := `INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES (:user_id, :token_hash, :expires_at)
		RETURNING id, created_at`
	rows, err := r.db.NamedQueryContext(ctx, query, s)
	if err != nil {
		return translateError(err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&s.ID, &s.CreatedAt)
	}
	return ErrNotFound
}

func (r *pgSessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (model.Session, error) {
	var s model.Session
	err := r.db.GetContext(ctx, &s,
		`SELECT * FROM sessions WHERE token_hash = $1 AND expires_at > NOW()`, tokenHash)
	return s, translateError(err)
}

func (r *pgSessionRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return translateError(err)
}

func (r *pgSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = $1`, userID)
	return translateError(err)
}

func (r *pgSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, translateError(err)
	}
	return result.RowsAffected()
}
