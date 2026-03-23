package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type OrgHandler struct {
	service service.OrgService
}

func NewOrgHandler(s service.OrgService) *OrgHandler {
	return &OrgHandler{service: s}
}

func (h *OrgHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/orgs", h.Create)
	mux.HandleFunc("GET /api/v1/orgs/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/orgs/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/orgs/{id}", h.Delete)
}

func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.OrgCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	org, err := h.service.Create(r.Context(), input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, org)
}

func (h *OrgHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	org, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.OrgUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	org, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
