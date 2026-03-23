package repository

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("duplicate record")
)

// translateError converts sql/pq errors into repository sentinels.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrConflict
	}
	return err
}
