package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/dukerupert/cairnpost/internal/view"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func parseOptionalUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func parseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func validationErrors(err error) view.FormErrors {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		return view.FormErrors{ve.Field: ve.Message}
	}
	return view.FormErrors{"": "An unexpected error occurred"}
}

// Option loaders for form dropdowns

func (h *PageHandler) companyOptions(ctx context.Context) []view.SelectOption {
	companies, _ := h.companies.List(ctx, h.orgID, repository.CompanyFilter{})
	opts := make([]view.SelectOption, 0, len(companies)+1)
	opts = append(opts, view.SelectOption{Value: "", Label: "— None —"})
	for _, c := range companies {
		opts = append(opts, view.SelectOption{Value: c.ID.String(), Label: c.Name})
	}
	return opts
}

func (h *PageHandler) contactOptions(ctx context.Context) []view.SelectOption {
	contacts, _ := h.contacts.List(ctx, h.orgID, repository.ContactFilter{})
	opts := make([]view.SelectOption, 0, len(contacts))
	for _, c := range contacts {
		name := c.FirstName
		if c.LastName != "" {
			name += " " + c.LastName
		}
		opts = append(opts, view.SelectOption{Value: c.ID.String(), Label: name})
	}
	return opts
}

func (h *PageHandler) userOptions(ctx context.Context) []view.SelectOption {
	users, _ := h.userRepo.List(ctx, h.orgID, repository.UserFilter{})
	opts := make([]view.SelectOption, 0, len(users))
	for _, u := range users {
		opts = append(opts, view.SelectOption{Value: u.ID.String(), Label: u.Name})
	}
	return opts
}

func (h *PageHandler) dealOptions(ctx context.Context) []view.SelectOption {
	deals, _ := h.deals.List(ctx, h.orgID, repository.DealFilter{})
	opts := make([]view.SelectOption, 0, len(deals)+1)
	opts = append(opts, view.SelectOption{Value: "", Label: "— None —"})
	for _, d := range deals {
		opts = append(opts, view.SelectOption{Value: d.ID.String(), Label: d.Title})
	}
	return opts
}

func stageOptions() []view.SelectOption {
	opts := make([]view.SelectOption, len(model.DefaultStages))
	for i, s := range model.DefaultStages {
		opts[i] = view.SelectOption{Value: s, Label: s}
	}
	return opts
}

func setSelected(opts []view.SelectOption, value string) []view.SelectOption {
	for i := range opts {
		opts[i].Selected = opts[i].Value == value
	}
	return opts
}

func uuidPtrToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
