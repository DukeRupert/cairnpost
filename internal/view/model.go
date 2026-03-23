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

// Page context for layout rendering
type PageContext struct {
	Title       string
	CurrentPath string
	UserName    string
}

// Form types

type SelectOption struct {
	Value    string
	Label    string
	Selected bool
}

type FormErrors map[string]string

func (e FormErrors) Get(field string) string {
	if e == nil {
		return ""
	}
	return e[field]
}

type ContactFormData struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Tags      string
	CompanyID string
}

type CompanyFormData struct {
	ID      string
	Name    string
	Address string
	Website string
	Notes   string
}

type DealFormData struct {
	ID        string
	Title     string
	Stage     string
	Value     string
	ContactID string
	CompanyID string
}

type TaskFormData struct {
	ID         string
	Title      string
	DueDate    string
	ContactID  string
	DealID     string
	AssignedTo string
}
