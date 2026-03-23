package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	OrgID     uuid.UUID `db:"org_id" json:"org_id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Role      Role      `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
