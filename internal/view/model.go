package view

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ContactRow struct {
	ID    uuid.UUID
	Name  string
	Email string
	Phone string
	Tags  []string
}

type CompanyRow struct {
	ID      uuid.UUID
	Name    string
	Website string
	Address string
}

type DealRow struct {
	ID          uuid.UUID
	Title       string
	Stage       string
	Value       decimal.Decimal
	ContactName string
}

type TaskRow struct {
	ID           uuid.UUID
	Title        string
	DueDate      *time.Time
	Done         bool
	AssignedName string
}

type ActivityRow struct {
	ID         uuid.UUID
	Type       string
	Body       string
	UserName   string
	OccurredAt time.Time
}

type ContactDetail struct {
	ID          uuid.UUID
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	Tags        []string
	CompanyName string
	CompanyID   *uuid.UUID
	CreatedAt   time.Time
}

type DealDetail struct {
	ID          uuid.UUID
	Title       string
	Stage       string
	Value       decimal.Decimal
	ContactName string
	ContactID   uuid.UUID
	CompanyName string
	CompanyID   *uuid.UUID
	ClosedAt    *time.Time
	CreatedAt   time.Time
}

type CompanyDetail struct {
	ID        uuid.UUID
	Name      string
	Address   string
	Website   string
	Notes     string
	CreatedAt time.Time
}
