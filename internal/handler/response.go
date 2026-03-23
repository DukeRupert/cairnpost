package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/google/uuid"
)

type envelope struct {
	Data       any             `json:"data,omitempty"`
	Pagination *paginationMeta `json:"pagination,omitempty"`
	Error      *errorBody      `json:"error,omitempty"`
}

type paginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type errorBody struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope{Data: data}); err != nil {
		log.Printf("respondJSON encode error: %v", err)
	}
}

func respondList(w http.ResponseWriter, data any, limit, offset, count int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := envelope{
		Data: data,
		Pagination: &paginationMeta{
			Limit:  limit,
			Offset: offset,
			Count:  count,
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("respondList encode error: %v", err)
	}
}

func respondError(w http.ResponseWriter, err error) {
	status, message := mapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(envelope{
		Error: &errorBody{Status: status, Message: message},
	}); encErr != nil {
		log.Printf("respondError encode error: %v", encErr)
	}
}

func mapError(err error) (int, string) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		return http.StatusUnprocessableEntity, ve.Error()
	}
	if errors.Is(err, repository.ErrNotFound) {
		return http.StatusNotFound, "not found"
	}
	if errors.Is(err, repository.ErrConflict) {
		return http.StatusConflict, "duplicate record"
	}
	log.Printf("internal error: %v", err)
	return http.StatusInternalServerError, "internal server error"
}

// decodeJSON reads a JSON request body into dst.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// Query param helpers

func queryString(r *http.Request, key string) *string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	return &v
}

func queryUUID(r *http.Request, key string) *uuid.UUID {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil
	}
	return &id
}

func queryBool(r *http.Request, key string) *bool {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func queryTime(r *http.Request, key string) *time.Time {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}

func queryPagination(r *http.Request) repository.Pagination {
	p := repository.Pagination{}
	if v := r.URL.Query().Get("limit"); v != "" {
		p.Limit, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		p.Offset, _ = strconv.Atoi(v)
	}
	return p
}
