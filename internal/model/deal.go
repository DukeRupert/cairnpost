package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Default pipeline stages — user-configurable per org in v2.
var DefaultStages = []string{
	"New Lead",
	"Estimate Sent",
	"Follow-up",
	"Won",
	"Lost",
}

type Deal struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	OrgID     uuid.UUID       `db:"org_id" json:"org_id"`
	Title     string          `db:"title" json:"title"`
	Stage     string          `db:"stage" json:"stage"`
	Value     decimal.Decimal `db:"value" json:"value"`
	ContactID uuid.UUID       `db:"contact_id" json:"contact_id"`
	CompanyID *uuid.UUID      `db:"company_id" json:"company_id"`
	ClosedAt  *time.Time      `db:"closed_at" json:"closed_at"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}
