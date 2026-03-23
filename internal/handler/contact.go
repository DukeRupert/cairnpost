package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type ContactHandler struct {
	service service.ContactService
}

func NewContactHandler(s service.ContactService) *ContactHandler {
	return &ContactHandler{service: s}
}

func (h *ContactHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("GET "+prefix+"/contacts", h.List)
	mux.HandleFunc("POST "+prefix+"/contacts", h.Create)
	mux.HandleFunc("GET "+prefix+"/contacts/{id}", h.Get)
	mux.HandleFunc("PUT "+prefix+"/contacts/{id}", h.Update)
	mux.HandleFunc("DELETE "+prefix+"/contacts/{id}", h.Delete)
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	filter := repository.ContactFilter{
		Search:     queryString(r, "search"),
		Tag:        queryString(r, "tag"),
		CompanyID:  queryUUID(r, "company_id"),
		Pagination: queryPagination(r),
	}

	contacts, err := h.service.List(r.Context(), org.ID, filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondList(w, contacts, filter.Pagination.Limit, filter.Pagination.Offset, len(contacts))
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var input service.ContactCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	contact, err := h.service.Create(r.Context(), org.ID, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, contact)
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	contact, err := h.service.GetByID(r.Context(), org.ID, id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, contact)
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.ContactUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	contact, err := h.service.Update(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, contact)
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	if err := h.service.Delete(r.Context(), org.ID, id); err != nil {
		respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
