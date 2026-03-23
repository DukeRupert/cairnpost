package model

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OrgID     uuid.UUID `db:"org_id" json:"org_id"`
	Name      string    `db:"name" json:"name"`
	Address   string    `db:"address" json:"address"`
	Website   string    `db:"website" json:"website"`
	Notes     string    `db:"notes" json:"notes"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
