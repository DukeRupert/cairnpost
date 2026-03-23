package handler

import (
	"net/http"

	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type TaskHandler struct {
	service service.TaskService
}

func NewTaskHandler(s service.TaskService) *TaskHandler {
	return &TaskHandler{service: s}
}

func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc("GET "+prefix+"/tasks", h.List)
	mux.HandleFunc("POST "+prefix+"/tasks", h.Create)
	mux.HandleFunc("GET "+prefix+"/tasks/{id}", h.Get)
	mux.HandleFunc("PUT "+prefix+"/tasks/{id}", h.Update)
	mux.HandleFunc("PATCH "+prefix+"/tasks/{id}/done", h.MarkDone)
	mux.HandleFunc("DELETE "+prefix+"/tasks/{id}", h.Delete)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	filter := repository.TaskFilter{
		AssignedTo: queryUUID(r, "assigned_to"),
		ContactID:  queryUUID(r, "contact_id"),
		DealID:     queryUUID(r, "deal_id"),
		Done:       queryBool(r, "done"),
		DueBefore:  queryTime(r, "due_before"),
		DueAfter:   queryTime(r, "due_after"),
		Overdue:    queryBool(r, "overdue"),
		Pagination: queryPagination(r),
	}

	tasks, err := h.service.List(r.Context(), org.ID, filter)
	if err != nil {
		respondError(w, err)
		return
	}
	respondList(w, tasks, filter.Pagination.Limit, filter.Pagination.Offset, len(tasks))
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())

	var input service.TaskCreateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	task, err := h.service.Create(r.Context(), org.ID, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	task, err := h.service.GetByID(r.Context(), org.ID, id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.TaskUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	task, err := h.service.Update(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) MarkDone(w http.ResponseWriter, r *http.Request) {
	org := OrgFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, &service.ValidationError{Field: "id", Message: "invalid UUID"})
		return
	}

	var input service.TaskDoneInput
	if err := decodeJSON(r, &input); err != nil {
		respondError(w, &service.ValidationError{Field: "body", Message: "invalid JSON"})
		return
	}

	task, err := h.service.MarkDone(r.Context(), org.ID, id, input)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
