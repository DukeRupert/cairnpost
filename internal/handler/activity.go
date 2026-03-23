package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type ActivityHandler struct {
	service service.ActivityService
}

func NewActivityHandler(s service.ActivityService) *ActivityHandler {
	return &ActivityHandler{service: s}
}

func (h *ActivityHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("GET "+prefix+"/activities", h.List)
	mux.HandleFunc("POST "+prefix+"/activities", h.Create)
	mux.HandleFunc("GET "+prefix+"/activities/{id}", h.Get)
}

func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var typeFilter *model.ActivityType
	if v := queryString(r, "type"); v != nil {
		t := model.ActivityType(*v)
		typeFilter = &t
	}

	filter := repository.ActivityFilter{
		ContactID:  queryUUID(r, "contact_id"),
		DealID:     queryUUID(r, "deal_id"),
		Type:       typeFilter,
		Pagination: queryPagination(r),
	}

	activities, err := h.service.List(r.Context(), org.ID, filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondList(w, activities, filter.Pagination.Limit, filter.Pagination.Offset, len(activities))
}

func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var input service.ActivityCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	activity, err := h.service.Create(r.Context(), org.ID, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, activity)
}

func (h *ActivityHandler) Get(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	activity, err := h.service.GetByID(r.Context(), org.ID, id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, activity)
}
