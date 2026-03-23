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
