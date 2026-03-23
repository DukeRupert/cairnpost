package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/web/templates/pages"
)

type PageHandler struct {
	orgSlug string
}

func NewPageHandler(orgSlug string) *PageHandler {
	return &PageHandler{orgSlug: orgSlug}
}

func (h *PageHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.Today)
	mux.HandleFunc("GET /contacts", h.Contacts)
	mux.HandleFunc("GET /companies", h.Companies)
	mux.HandleFunc("GET /deals", h.Deals)
	mux.HandleFunc("GET /tasks", h.Tasks)
	mux.HandleFunc("GET /settings", h.Settings)
}

func (h *PageHandler) Today(w http.ResponseWriter, r *http.Request) {
	pages.Today().Render(r.Context(), w)
}

func (h *PageHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	pages.Contacts().Render(r.Context(), w)
}

func (h *PageHandler) Companies(w http.ResponseWriter, r *http.Request) {
	pages.Companies().Render(r.Context(), w)
}

func (h *PageHandler) Deals(w http.ResponseWriter, r *http.Request) {
	pages.Deals().Render(r.Context(), w)
}

func (h *PageHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	pages.Tasks().Render(r.Context(), w)
}

func (h *PageHandler) Settings(w http.ResponseWriter, r *http.Request) {
	pages.Settings().Render(r.Context(), w)
}
