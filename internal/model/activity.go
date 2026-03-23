package model

import (
	"time"

	"github.com/google/uuid"
)

type ActivityType string

const (
	ActivityNote      ActivityType = "note"
	ActivityCall      ActivityType = "call"
	ActivityEmail     ActivityType = "email"
	ActivitySMS       ActivityType = "sms"
	ActivitySiteVisit ActivityType = "site_visit"
)

type Activity struct {
	ID         uuid.UUID    `db:"id" json:"id"`
	OrgID      uuid.UUID    `db:"org_id" json:"org_id"`
	Type       ActivityType `db:"type" json:"type"`
	Body       string       `db:"body" json:"body"`
	ContactID  uuid.UUID    `db:"contact_id" json:"contact_id"`
	DealID     *uuid.UUID   `db:"deal_id" json:"deal_id"`
	UserID     uuid.UUID    `db:"user_id" json:"user_id"`
	OccurredAt time.Time    `db:"occurred_at" json:"occurred_at"`
	CreatedAt  time.Time    `db:"created_at" json:"created_at"`
}
