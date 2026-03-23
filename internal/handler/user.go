package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("GET "+prefix+"/users", h.List)
	mux.HandleFunc("POST "+prefix+"/users", h.Create)
	mux.HandleFunc("GET "+prefix+"/users/{id}", h.Get)
	mux.HandleFunc("PUT "+prefix+"/users/{id}", h.Update)
	mux.HandleFunc("DELETE "+prefix+"/users/{id}", h.Delete)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var roleFilter *model.Role
	if v := queryString(r, "role"); v != nil {
		role := model.Role(*v)
		roleFilter = &role
	}

	filter := repository.UserFilter{
		Role:       roleFilter,
		Pagination: queryPagination(r),
	}

	users, err := h.service.List(r.Context(), org.ID, filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondList(w, users, filter.Pagination.Limit, filter.Pagination.Offset, len(users))
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var input service.UserCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	user, err := h.service.Create(r.Context(), org.ID, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	user, err := h.service.GetByID(r.Context(), org.ID, id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.UserUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	user, err := h.service.Update(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
