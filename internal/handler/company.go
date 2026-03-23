package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	service service.CompanyService
}

func NewCompanyHandler(s service.CompanyService) *CompanyHandler {
	return &CompanyHandler{service: s}
}

func (h *CompanyHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("GET "+prefix+"/companies", h.List)
	mux.HandleFunc("POST "+prefix+"/companies", h.Create)
	mux.HandleFunc("GET "+prefix+"/companies/{id}", h.Get)
	mux.HandleFunc("PUT "+prefix+"/companies/{id}", h.Update)
	mux.HandleFunc("DELETE "+prefix+"/companies/{id}", h.Delete)
}

func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	filter := repository.CompanyFilter{
		Search:     queryString(r, "search"),
		Pagination: queryPagination(r),
	}

	companies, err := h.service.List(r.Context(), org.ID, filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondList(w, companies, filter.Pagination.Limit, filter.Pagination.Offset, len(companies))
}

func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var input service.CompanyCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	company, err := h.service.Create(r.Context(), org.ID, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, company)
}

func (h *CompanyHandler) Get(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	company, err := h.service.GetByID(r.Context(), org.ID, id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.CompanyUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	company, err := h.service.Update(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
