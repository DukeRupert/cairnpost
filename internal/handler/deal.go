package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type DealHandler struct {
	service service.DealService
}

func NewDealHandler(s service.DealService) *DealHandler {
	return &DealHandler{service: s}
}

func (h *DealHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("GET "+prefix+"/deals", h.List)
	mux.HandleFunc("POST "+prefix+"/deals", h.Create)
	mux.HandleFunc("GET "+prefix+"/deals/{id}", h.Get)
	mux.HandleFunc("PUT "+prefix+"/deals/{id}", h.Update)
	mux.HandleFunc("PATCH "+prefix+"/deals/{id}/stage", h.UpdateStage)
	mux.HandleFunc("DELETE "+prefix+"/deals/{id}", h.Delete)
}

func (h *DealHandler) List(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	filter := repository.DealFilter{
		Stage:      queryString(r, "stage"),
		ContactID:  queryUUID(r, "contact_id"),
		CompanyID:  queryUUID(r, "company_id"),
		Open:       queryBool(r, "open"),
		Pagination: queryPagination(r),
	}

	deals, err := h.service.List(r.Context(), org.ID, filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondList(w, deals, filter.Pagination.Limit, filter.Pagination.Offset, len(deals))
}

func (h *DealHandler) Create(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var input service.DealCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	deal, err := h.service.Create(r.Context(), org.ID, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, deal)
}

func (h *DealHandler) Get(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	deal, err := h.service.GetByID(r.Context(), org.ID, id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, deal)
}

func (h *DealHandler) Update(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.DealUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	deal, err := h.service.Update(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, deal)
}

func (h *DealHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.DealStageInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	deal, err := h.service.UpdateStage(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, deal)
}

func (h *DealHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
