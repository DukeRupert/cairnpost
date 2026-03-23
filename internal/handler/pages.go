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
	"github.com/dukerupert/cairnpost/web/templates/components"
	"github.com/dukerupert/cairnpost/web/templates/pages"
	"github.com/google/uuid"
)

type PageHandler struct {
	orgID         uuid.UUID
	defaultUserID uuid.UUID
	contacts      service.ContactService
	companies     service.CompanyService
	deals         service.DealService
	tasks         service.TaskService
	activities    service.ActivityService
	contactRepo   repository.ContactRepository
	userRepo      repository.UserRepository
	companyRepo   repository.CompanyRepository
}

func NewPageHandler(
	orgID uuid.UUID,
	defaultUserID uuid.UUID,
	contacts service.ContactService,
	companies service.CompanyService,
	deals service.DealService,
	tasks service.TaskService,
	activities service.ActivityService,
	contactRepo repository.ContactRepository,
	userRepo repository.UserRepository,
	companyRepo repository.CompanyRepository,
) *PageHandler {
	return &PageHandler{
		orgID:         orgID,
		defaultUserID: defaultUserID,
		contacts:      contacts,
		companies:     companies,
		deals:         deals,
		tasks:         tasks,
		activities:    activities,
		contactRepo:   contactRepo,
		userRepo:      userRepo,
		companyRepo:   companyRepo,
	}
}

func (h *PageHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.Today)

	mux.HandleFunc("GET /contacts", h.Contacts)
	mux.HandleFunc("GET /contacts/new", h.ContactNew)
	mux.HandleFunc("POST /contacts", h.ContactCreate)
	mux.HandleFunc("GET /contacts/{id}", h.ContactDetail)
	mux.HandleFunc("GET /contacts/{id}/edit", h.ContactEdit)
	mux.HandleFunc("POST /contacts/{id}", h.ContactUpdate)
	mux.HandleFunc("POST /contacts/{id}/activities", h.CreateContactActivity)

	mux.HandleFunc("GET /companies", h.Companies)
	mux.HandleFunc("GET /companies/new", h.CompanyNew)
	mux.HandleFunc("POST /companies", h.CompanyCreate)
	mux.HandleFunc("GET /companies/{id}", h.CompanyDetail)
	mux.HandleFunc("GET /companies/{id}/edit", h.CompanyEdit)
	mux.HandleFunc("POST /companies/{id}", h.CompanyUpdate)

	mux.HandleFunc("GET /deals", h.Deals)
	mux.HandleFunc("GET /deals/new", h.DealNew)
	mux.HandleFunc("POST /deals", h.DealCreate)
	mux.HandleFunc("GET /deals/{id}", h.DealDetail)
	mux.HandleFunc("GET /deals/{id}/edit", h.DealEdit)
	mux.HandleFunc("POST /deals/{id}", h.DealUpdate)
	mux.HandleFunc("POST /deals/{id}/activities", h.CreateDealActivity)

	mux.HandleFunc("GET /tasks", h.Tasks)
	mux.HandleFunc("GET /tasks/new", h.TaskNew)
	mux.HandleFunc("POST /tasks", h.TaskCreate)
	mux.HandleFunc("GET /tasks/{id}/edit", h.TaskEdit)
	mux.HandleFunc("POST /tasks/{id}", h.TaskUpdate)

	mux.HandleFunc("GET /settings", h.Settings)
}

// --- List pages ---

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

	rows := h.contactRows(models)

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

	rows := h.dealRows(ctx, models)
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

// --- Detail pages ---

func (h *PageHandler) ContactDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	contact, err := h.contacts.GetByID(ctx, h.orgID, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	companyName := ""
	if contact.CompanyID != nil {
		if co, err := h.companyRepo.GetByID(ctx, h.orgID, *contact.CompanyID); err == nil {
			companyName = co.Name
		}
	}

	detail := view.ContactDetail{
		ID:          contact.ID,
		FirstName:   contact.FirstName,
		LastName:    contact.LastName,
		Email:       contact.Email,
		Phone:       contact.Phone,
		Tags:        []string(contact.Tags),
		CompanyName: companyName,
		CompanyID:   contact.CompanyID,
		CreatedAt:   contact.CreatedAt,
	}

	activityModels, _ := h.activities.List(ctx, h.orgID, repository.ActivityFilter{ContactID: &id})
	activities := h.activityRows(ctx, activityModels)

	dealModels, _ := h.deals.List(ctx, h.orgID, repository.DealFilter{ContactID: &id})
	deals := h.dealRows(ctx, dealModels)

	taskModels, _ := h.tasks.List(ctx, h.orgID, repository.TaskFilter{ContactID: &id})
	tasks := h.taskRows(ctx, taskModels)

	pages.ContactDetailPage(detail, activities, deals, tasks).Render(ctx, w)
}

func (h *PageHandler) DealDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	deal, err := h.deals.GetByID(ctx, h.orgID, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	contactName := ""
	if c, err := h.contactRepo.GetByID(ctx, h.orgID, deal.ContactID); err == nil {
		contactName = c.FirstName
		if c.LastName != "" {
			contactName += " " + c.LastName
		}
	}

	companyName := ""
	if deal.CompanyID != nil {
		if co, err := h.companyRepo.GetByID(ctx, h.orgID, *deal.CompanyID); err == nil {
			companyName = co.Name
		}
	}

	detail := view.DealDetail{
		ID:          deal.ID,
		Title:       deal.Title,
		Stage:       deal.Stage,
		Value:       deal.Value,
		ContactName: contactName,
		ContactID:   deal.ContactID,
		CompanyName: companyName,
		CompanyID:   deal.CompanyID,
		ClosedAt:    deal.ClosedAt,
		CreatedAt:   deal.CreatedAt,
	}

	activityModels, _ := h.activities.List(ctx, h.orgID, repository.ActivityFilter{DealID: &id})
	activities := h.activityRows(ctx, activityModels)

	taskModels, _ := h.tasks.List(ctx, h.orgID, repository.TaskFilter{DealID: &id})
	tasks := h.taskRows(ctx, taskModels)

	pages.DealDetailPage(detail, activities, tasks).Render(ctx, w)
}

func (h *PageHandler) CompanyDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	company, err := h.companies.GetByID(ctx, h.orgID, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	detail := view.CompanyDetail{
		ID:        company.ID,
		Name:      company.Name,
		Address:   company.Address,
		Website:   company.Website,
		Notes:     company.Notes,
		CreatedAt: company.CreatedAt,
	}

	contactModels, _ := h.contacts.List(ctx, h.orgID, repository.ContactFilter{CompanyID: &id})
	contacts := h.contactRows(contactModels)

	dealModels, _ := h.deals.List(ctx, h.orgID, repository.DealFilter{CompanyID: &id})
	deals := h.dealRows(ctx, dealModels)

	pages.CompanyDetailPage(detail, contacts, deals).Render(ctx, w)
}

// --- Activity POST handlers ---

func (h *PageHandler) CreateContactActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	_, err = h.activities.Create(ctx, h.orgID, service.ActivityCreateInput{
		Type:      model.ActivityType(r.FormValue("type")),
		Body:      r.FormValue("body"),
		ContactID: contactID,
		UserID:    h.defaultUserID,
	})
	if err != nil {
		log.Printf("create contact activity: %v", err)
	}

	activityModels, _ := h.activities.List(ctx, h.orgID, repository.ActivityFilter{ContactID: &contactID})
	activities := h.activityRows(ctx, activityModels)
	components.ActivityTimeline(activities, contactID, nil).Render(ctx, w)
}

func (h *PageHandler) CreateDealActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dealID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deal, err := h.deals.GetByID(ctx, h.orgID, dealID)
	if err != nil {
		http.Error(w, "deal not found", http.StatusNotFound)
		return
	}

	r.ParseForm()
	_, err = h.activities.Create(ctx, h.orgID, service.ActivityCreateInput{
		Type:      model.ActivityType(r.FormValue("type")),
		Body:      r.FormValue("body"),
		ContactID: deal.ContactID,
		DealID:    &dealID,
		UserID:    h.defaultUserID,
	})
	if err != nil {
		log.Printf("create deal activity: %v", err)
	}

	activityModels, _ := h.activities.List(ctx, h.orgID, repository.ActivityFilter{DealID: &dealID})
	activities := h.activityRows(ctx, activityModels)
	components.ActivityTimeline(activities, deal.ContactID, &dealID).Render(ctx, w)
}

// --- Helpers ---

func (h *PageHandler) contactRows(models []model.Contact) []view.ContactRow {
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
	return rows
}

func (h *PageHandler) dealRows(ctx context.Context, models []model.Deal) []view.DealRow {
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
	return rows
}

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

func (h *PageHandler) activityRows(ctx context.Context, activities []model.Activity) []view.ActivityRow {
	userNames := make(map[uuid.UUID]string)
	for _, a := range activities {
		if _, ok := userNames[a.UserID]; !ok {
			u, err := h.userRepo.GetByID(ctx, h.orgID, a.UserID)
			if err == nil {
				userNames[a.UserID] = u.Name
			}
		}
	}

	rows := make([]view.ActivityRow, len(activities))
	for i, a := range activities {
		rows[i] = view.ActivityRow{
			ID:         a.ID,
			Type:       string(a.Type),
			Body:       a.Body,
			UserName:   userNames[a.UserID],
			OccurredAt: a.OccurredAt,
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
