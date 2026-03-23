package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/dukerupert/cairnpost/internal/auth"
	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/dukerupert/cairnpost/web/templates/pages"
	"github.com/google/uuid"
)

const (
	sessionCookieName = "cairnpost_session"
	sessionDuration   = 30 * 24 * time.Hour // 30 days
)

type AuthHandler struct {
	orgID        uuid.UUID
	userSvc      service.UserService
	sessionRepo  repository.SessionRepository
	secureCookie bool
}

func NewAuthHandler(
	orgID uuid.UUID,
	userSvc service.UserService,
	sessionRepo repository.SessionRepository,
	secureCookie bool,
) *AuthHandler {
	return &AuthHandler{
		orgID:        orgID,
		userSvc:      userSvc,
		sessionRepo:  sessionRepo,
		secureCookie: secureCookie,
	}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.LoginPage("").Render(r.Context(), w)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.userSvc.Authenticate(ctx, h.orgID, email, password)
	if err != nil {
		pages.LoginPage("Invalid email or password.").Render(ctx, w)
		return
	}

	raw, hash, err := auth.GenerateSessionToken()
	if err != nil {
		log.Printf("auth: generating session token: %v", err)
		pages.LoginPage("An error occurred. Please try again.").Render(ctx, w)
		return
	}

	session := model.Session{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.sessionRepo.Create(ctx, &session); err != nil {
		log.Printf("auth: creating session: %v", err)
		pages.LoginPage("An error occurred. Please try again.").Render(ctx, w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    raw,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		tokenHash := auth.HashToken(cookie.Value)
		h.sessionRepo.DeleteByTokenHash(r.Context(), tokenHash)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
