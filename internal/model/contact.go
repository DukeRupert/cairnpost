package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Contact struct {
	ID        uuid.UUID      `db:"id" json:"id"`
	OrgID     uuid.UUID      `db:"org_id" json:"org_id"`
	FirstName string         `db:"first_name" json:"first_name"`
	LastName  string         `db:"last_name" json:"last_name"`
	Email     string         `db:"email" json:"email"`
	Phone     string         `db:"phone" json:"phone"`
	Tags      pq.StringArray `db:"tags" json:"tags"`
	CompanyID *uuid.UUID     `db:"company_id" json:"company_id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}
