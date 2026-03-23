package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/dukerupert/cairnpost/internal/view"
	"github.com/dukerupert/cairnpost/web/templates/pages"
	"github.com/google/uuid"
)

type PageHandler struct {
	orgID       uuid.UUID
	contacts    service.ContactService
	companies   service.CompanyService
	deals       service.DealService
	tasks       service.TaskService
	contactRepo repository.ContactRepository
	userRepo    repository.UserRepository
}

func NewPageHandler(
	orgID uuid.UUID,
	contacts service.ContactService,
	companies service.CompanyService,
	deals service.DealService,
	tasks service.TaskService,
	contactRepo repository.ContactRepository,
	userRepo repository.UserRepository,
) *PageHandler {
	return &PageHandler{
		orgID:       orgID,
		contacts:    contacts,
		companies:   companies,
		deals:       deals,
		tasks:       tasks,
		contactRepo: contactRepo,
		userRepo:    userRepo,
	}
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
	ctx := r.Context()
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)

	overdueModels, err := h.tasks.List(ctx, h.orgID, repository.TaskFilter{
		Overdue: boolPtr(true),
		Done:    boolPtr(false),
	})
	if err != nil {
		log.Printf("today: loading overdue tasks: %v", err)
	}

	dueTodayModels, err := h.tasks.List(ctx, h.orgID, repository.TaskFilter{
		DueAfter:  &startOfDay,
		DueBefore: &endOfDay,
		Done:      boolPtr(false),
	})
	if err != nil {
		log.Printf("today: loading due-today tasks: %v", err)
	}

	overdue := h.taskRows(ctx, overdueModels)
	dueToday := h.taskRows(ctx, dueTodayModels)

	pages.TodayPage(overdue, dueToday).Render(ctx, w)
}

func (h *PageHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := repository.ContactFilter{
		Search:     queryString(r, "search"),
		Pagination: queryPagination(r),
	}

	models, err := h.contacts.List(ctx, h.orgID, filter)
	if err != nil {
		log.Printf("contacts: %v", err)
	}

	rows := make([]view.ContactRow, len(models))
	for i, c := range models {
		name := c.FirstName
		if c.LastName != "" {
			name += " " + c.LastName
		}
		rows[i] = view.ContactRow{
			ID:    c.ID,
			Name:  name,
			Email: c.Email,
			Phone: c.Phone,
			Tags:  []string(c.Tags),
		}
	}

	search := ""
	if filter.Search != nil {
		search = *filter.Search
	}

	if isHTMX(r) {
		pages.ContactsTable(rows, search).Render(ctx, w)
		return
	}
	pages.ContactsPage(rows, search).Render(ctx, w)
}

func (h *PageHandler) Companies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := repository.CompanyFilter{
		Search:     queryString(r, "search"),
		Pagination: queryPagination(r),
	}

	models, err := h.companies.List(ctx, h.orgID, filter)
	if err != nil {
		log.Printf("companies: %v", err)
	}

	rows := make([]view.CompanyRow, len(models))
	for i, c := range models {
		rows[i] = view.CompanyRow{
			ID:      c.ID,
			Name:    c.Name,
			Website: c.Website,
			Address: c.Address,
		}
	}

	search := ""
	if filter.Search != nil {
		search = *filter.Search
	}

	if isHTMX(r) {
		pages.CompaniesTable(rows, search).Render(ctx, w)
		return
	}
	pages.CompaniesPage(rows, search).Render(ctx, w)
}

func (h *PageHandler) Deals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := repository.DealFilter{
		Stage:      queryString(r, "stage"),
		Open:       queryBool(r, "open"),
		Pagination: queryPagination(r),
	}

	models, err := h.deals.List(ctx, h.orgID, filter)
	if err != nil {
		log.Printf("deals: %v", err)
	}

	// Resolve contact names
	contactNames := make(map[uuid.UUID]string)
	for _, d := range models {
		if _, ok := contactNames[d.ContactID]; !ok {
			c, err := h.contactRepo.GetByID(ctx, h.orgID, d.ContactID)
			if err == nil {
				name := c.FirstName
				if c.LastName != "" {
					name += " " + c.LastName
				}
				contactNames[d.ContactID] = name
			}
		}
	}

	rows := make([]view.DealRow, len(models))
	for i, d := range models {
		rows[i] = view.DealRow{
			ID:          d.ID,
			Title:       d.Title,
			Stage:       d.Stage,
			Value:       d.Value,
			ContactName: contactNames[d.ContactID],
		}
	}

	pages.DealsPage(rows).Render(ctx, w)
}

func (h *PageHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := repository.TaskFilter{
		Done:       queryBool(r, "done"),
		Overdue:    queryBool(r, "overdue"),
		Pagination: queryPagination(r),
	}

	models, err := h.tasks.List(ctx, h.orgID, filter)
	if err != nil {
		log.Printf("tasks: %v", err)
	}

	rows := h.taskRows(ctx, models)
	pages.TasksPage(rows).Render(ctx, w)
}

func (h *PageHandler) Settings(w http.ResponseWriter, r *http.Request) {
	pages.Settings().Render(r.Context(), w)
}

// taskRows converts model.Task slice to view.TaskRow slice, resolving assignee names.
func (h *PageHandler) taskRows(ctx context.Context, tasks []model.Task) []view.TaskRow {
	userNames := make(map[uuid.UUID]string)
	for _, t := range tasks {
		if _, ok := userNames[t.AssignedTo]; !ok {
			u, err := h.userRepo.GetByID(ctx, h.orgID, t.AssignedTo)
			if err == nil {
				userNames[t.AssignedTo] = u.Name
			}
		}
	}

	rows := make([]view.TaskRow, len(tasks))
	for i, t := range tasks {
		rows[i] = view.TaskRow{
			ID:           t.ID,
			Title:        t.Title,
			DueDate:      t.DueDate,
			Done:         t.Done,
			AssignedName: userNames[t.AssignedTo],
		}
	}
	return rows
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func boolPtr(b bool) *bool {
	return &b
}
