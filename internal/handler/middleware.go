package handler

import (
	"context"
	"net/http"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
)

type contextKey string

const orgContextKey contextKey = "org"

// OrgFromContext retrieves the model.Org stored by OrgResolver.
func OrgFromContext(ctx context.Context) model.Org {
	return ctx.Value(orgContextKey).(model.Org)
}

// OrgResolver returns middleware that resolves the {org} slug from the URL path
// to a model.Org via the repository. If PathValue("org") is empty (routes without
// {org}), the request passes through unchanged.
func OrgResolver(orgRepo repository.OrgRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := r.PathValue("org")
			if slug == "" {
				next.ServeHTTP(w, r)
				return
			}

			org, err := orgRepo.GetBySlug(r.Context(), slug)
			if err != nil {
				respondError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), orgContextKey, org)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
