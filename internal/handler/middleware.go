package handler

import (
	"context"
	"net/http"

	"github.com/dukerupert/cairnpost/internal/auth"
	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
)

type contextKey string

const orgContextKey contextKey = "org"

// OrgFromContext retrieves the model.Org stored by OrgResolver.
func OrgFromContext(ctx context.Context) model.Org {
	return ctx.Value(orgContextKey).(model.Org)
}

// RequireAuth returns middleware that checks for a valid session cookie.
// If invalid, redirects to /login (or sends HX-Redirect for htmx requests).
func RequireAuth(sessions repository.SessionRepository, users repository.UserRepository, orgID uuid.UUID, secureCookie bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil || cookie.Value == "" {
				redirectToLogin(w, r)
				return
			}

			tokenHash := auth.HashToken(cookie.Value)
			session, err := sessions.GetByTokenHash(r.Context(), tokenHash)
			if err != nil {
				clearSessionCookie(w, secureCookie)
				redirectToLogin(w, r)
				return
			}

			user, err := users.GetByID(r.Context(), orgID, session.UserID)
			if err != nil {
				sessions.DeleteByTokenHash(r.Context(), tokenHash)
				clearSessionCookie(w, secureCookie)
				redirectToLogin(w, r)
				return
			}

			ctx := auth.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
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
