package model

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	OrgID      uuid.UUID  `db:"org_id" json:"org_id"`
	Title      string     `db:"title" json:"title"`
	DueDate    *time.Time `db:"due_date" json:"due_date"`
	Done       bool       `db:"done" json:"done"`
	ContactID  *uuid.UUID `db:"contact_id" json:"contact_id"`
	DealID     *uuid.UUID `db:"deal_id" json:"deal_id"`
	AssignedTo uuid.UUID  `db:"assigned_to" json:"assigned_to"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}
