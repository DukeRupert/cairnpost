package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dukerupert/cairnpost/internal/service"
	"github.com/dukerupert/cairnpost/internal/view"
	"github.com/dukerupert/cairnpost/web/templates/pages"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func formTitle(entity, id string) string {
	if id == "" {
		return "New " + entity
	}
	return "Edit " + entity
}

// --- Company forms ---

func (h *PageHandler) CompanyNew(w http.ResponseWriter, r *http.Request) {
	data := view.CompanyFormData{}
	pages.CompanyFormPage(h.pc(r, "New Company", "/companies"), data, nil).Render(r.Context(), w)
}

func (h *PageHandler) CompanyCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	data := view.CompanyFormData{
		Name:    r.FormValue("name"),
		Address: r.FormValue("address"),
		Website: r.FormValue("website"),
		Notes:   r.FormValue("notes"),
	}

	company, err := h.companies.Create(ctx, h.orgID, service.CompanyCreateInput{
		Name:    data.Name,
		Address: data.Address,
		Website: data.Website,
		Notes:   data.Notes,
	})
	if err != nil {
		pages.CompanyFormPage(h.pc(r, "New Company", "/companies"), data, validationErrors(err)).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/companies/%s", company.ID), http.StatusSeeOther)
}

func (h *PageHandler) CompanyEdit(w http.ResponseWriter, r *http.Request) {
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

	data := view.CompanyFormData{
		ID:      company.ID.String(),
		Name:    company.Name,
		Address: company.Address,
		Website: company.Website,
		Notes:   company.Notes,
	}

	pages.CompanyFormPage(h.pc(r, "Edit Company", "/companies"), data, nil).Render(ctx, w)
}

func (h *PageHandler) CompanyUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()

	data := view.CompanyFormData{
		ID:      id.String(),
		Name:    r.FormValue("name"),
		Address: r.FormValue("address"),
		Website: r.FormValue("website"),
		Notes:   r.FormValue("notes"),
	}

	_, err = h.companies.Update(ctx, h.orgID, id, service.CompanyUpdateInput{
		Name:    strPtr(data.Name),
		Address: &data.Address,
		Website: &data.Website,
		Notes:   &data.Notes,
	})
	if err != nil {
		pages.CompanyFormPage(h.pc(r, "Edit Company", "/companies"), data, validationErrors(err)).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/companies/%s", id), http.StatusSeeOther)
}

// --- Contact forms ---

func (h *PageHandler) ContactNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companies := h.companyOptions(ctx)
	pages.ContactFormPage(h.pc(r, "New Contact", "/contacts"), view.ContactFormData{}, nil, companies).Render(ctx, w)
}

func (h *PageHandler) ContactCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	data := view.ContactFormData{
		FirstName: r.FormValue("first_name"),
		LastName:  r.FormValue("last_name"),
		Email:     r.FormValue("email"),
		Phone:     r.FormValue("phone"),
		Tags:      r.FormValue("tags"),
		CompanyID: r.FormValue("company_id"),
	}

	contact, err := h.contacts.Create(ctx, h.orgID, service.ContactCreateInput{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Phone:     data.Phone,
		Tags:      parseTags(data.Tags),
		CompanyID: parseOptionalUUID(data.CompanyID),
	})
	if err != nil {
		companies := setSelected(h.companyOptions(ctx), data.CompanyID)
		pages.ContactFormPage(h.pc(r, "New Contact", "/contacts"), data, validationErrors(err), companies).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/%s", contact.ID), http.StatusSeeOther)
}

func (h *PageHandler) ContactEdit(w http.ResponseWriter, r *http.Request) {
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

	data := view.ContactFormData{
		ID:        contact.ID.String(),
		FirstName: contact.FirstName,
		LastName:  contact.LastName,
		Email:     contact.Email,
		Phone:     contact.Phone,
		Tags:      strings.Join([]string(pq.StringArray(contact.Tags)), ", "),
		CompanyID: uuidPtrToString(contact.CompanyID),
	}

	companies := setSelected(h.companyOptions(ctx), data.CompanyID)
	pages.ContactFormPage(h.pc(r, "Edit Contact", "/contacts"), data, nil, companies).Render(ctx, w)
}

func (h *PageHandler) ContactUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()

	data := view.ContactFormData{
		ID:        id.String(),
		FirstName: r.FormValue("first_name"),
		LastName:  r.FormValue("last_name"),
		Email:     r.FormValue("email"),
		Phone:     r.FormValue("phone"),
		Tags:      r.FormValue("tags"),
		CompanyID: r.FormValue("company_id"),
	}

	tags := parseTags(data.Tags)
	_, err = h.contacts.Update(ctx, h.orgID, id, service.ContactUpdateInput{
		FirstName: strPtr(data.FirstName),
		LastName:  &data.LastName,
		Email:     &data.Email,
		Phone:     &data.Phone,
		Tags:      &tags,
		CompanyID: parseOptionalUUID(data.CompanyID),
	})
	if err != nil {
		companies := setSelected(h.companyOptions(ctx), data.CompanyID)
		pages.ContactFormPage(h.pc(r, "Edit Contact", "/contacts"), data, validationErrors(err), companies).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/contacts/%s", id), http.StatusSeeOther)
}

// --- Deal forms ---

func (h *PageHandler) DealNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := view.DealFormData{
		ContactID: r.URL.Query().Get("contact_id"),
	}
	contacts := setSelected(h.contactOptions(ctx), data.ContactID)
	companies := h.companyOptions(ctx)
	stages := stageOptions()
	pages.DealFormPage(h.pc(r, "New Deal", "/deals"), data, nil, contacts, companies, stages).Render(ctx, w)
}

func (h *PageHandler) DealCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	data := view.DealFormData{
		Title:     r.FormValue("title"),
		Stage:     r.FormValue("stage"),
		Value:     r.FormValue("value"),
		ContactID: r.FormValue("contact_id"),
		CompanyID: r.FormValue("company_id"),
	}

	contactID, err := uuid.Parse(data.ContactID)
	if err != nil {
		errs := view.FormErrors{"contact_id": "is required"}
		contacts := setSelected(h.contactOptions(ctx), data.ContactID)
		companies := setSelected(h.companyOptions(ctx), data.CompanyID)
		stages := setSelected(stageOptions(), data.Stage)
		pages.DealFormPage(h.pc(r, "New Deal", "/deals"), data, errs, contacts, companies, stages).Render(ctx, w)
		return
	}

	deal, err := h.deals.Create(ctx, h.orgID, service.DealCreateInput{
		Title:     data.Title,
		Stage:     data.Stage,
		Value:     parseDecimal(data.Value),
		ContactID: contactID,
		CompanyID: parseOptionalUUID(data.CompanyID),
	})
	if err != nil {
		contacts := setSelected(h.contactOptions(ctx), data.ContactID)
		companies := setSelected(h.companyOptions(ctx), data.CompanyID)
		stages := setSelected(stageOptions(), data.Stage)
		pages.DealFormPage(h.pc(r, "New Deal", "/deals"), data, validationErrors(err), contacts, companies, stages).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/deals/%s", deal.ID), http.StatusSeeOther)
}

func (h *PageHandler) DealEdit(w http.ResponseWriter, r *http.Request) {
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

	data := view.DealFormData{
		ID:        deal.ID.String(),
		Title:     deal.Title,
		Stage:     deal.Stage,
		Value:     deal.Value.StringFixed(2),
		ContactID: deal.ContactID.String(),
		CompanyID: uuidPtrToString(deal.CompanyID),
	}

	contacts := setSelected(h.contactOptions(ctx), data.ContactID)
	companies := setSelected(h.companyOptions(ctx), data.CompanyID)
	stages := setSelected(stageOptions(), data.Stage)
	pages.DealFormPage(h.pc(r, "Edit Deal", "/deals"), data, nil, contacts, companies, stages).Render(ctx, w)
}

func (h *PageHandler) DealUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()

	data := view.DealFormData{
		ID:        id.String(),
		Title:     r.FormValue("title"),
		Stage:     r.FormValue("stage"),
		Value:     r.FormValue("value"),
		ContactID: r.FormValue("contact_id"),
		CompanyID: r.FormValue("company_id"),
	}

	contactID := parseOptionalUUID(data.ContactID)
	val := parseDecimal(data.Value)
	_, err = h.deals.Update(ctx, h.orgID, id, service.DealUpdateInput{
		Title:     strPtr(data.Title),
		Stage:     strPtr(data.Stage),
		Value:     &val,
		ContactID: contactID,
		CompanyID: parseOptionalUUID(data.CompanyID),
	})
	if err != nil {
		contacts := setSelected(h.contactOptions(ctx), data.ContactID)
		companies := setSelected(h.companyOptions(ctx), data.CompanyID)
		stages := setSelected(stageOptions(), data.Stage)
		pages.DealFormPage(h.pc(r, "Edit Deal", "/deals"), data, validationErrors(err), contacts, companies, stages).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/deals/%s", id), http.StatusSeeOther)
}

// --- Task forms ---

func (h *PageHandler) TaskNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := view.TaskFormData{
		ContactID:  r.URL.Query().Get("contact_id"),
		DealID:     r.URL.Query().Get("deal_id"),
		AssignedTo: h.currentUser(r).ID.String(),
	}
	users := setSelected(h.userOptions(ctx), data.AssignedTo)
	contacts := setSelected(h.contactOptionsWithNone(ctx), data.ContactID)
	deals := setSelected(h.dealOptions(ctx), data.DealID)
	pages.TaskFormPage(h.pc(r, "New Task", "/tasks"), data, nil, users, contacts, deals).Render(ctx, w)
}

func (h *PageHandler) TaskCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	data := view.TaskFormData{
		Title:      r.FormValue("title"),
		DueDate:    r.FormValue("due_date"),
		ContactID:  r.FormValue("contact_id"),
		DealID:     r.FormValue("deal_id"),
		AssignedTo: r.FormValue("assigned_to"),
	}

	assignedTo, err := uuid.Parse(data.AssignedTo)
	if err != nil {
		errs := view.FormErrors{"assigned_to": "is required"}
		users := setSelected(h.userOptions(ctx), data.AssignedTo)
		contacts := setSelected(h.contactOptionsWithNone(ctx), data.ContactID)
		deals := setSelected(h.dealOptions(ctx), data.DealID)
		pages.TaskFormPage(h.pc(r, "New Task", "/tasks"), data, errs, users, contacts, deals).Render(ctx, w)
		return
	}

	input := service.TaskCreateInput{
		Title:      data.Title,
		ContactID:  parseOptionalUUID(data.ContactID),
		DealID:     parseOptionalUUID(data.DealID),
		AssignedTo: assignedTo,
	}
	if data.DueDate != "" {
		t, err := time.Parse("2006-01-02", data.DueDate)
		if err == nil {
			input.DueDate = &t
		}
	}

	_, err = h.tasks.Create(ctx, h.orgID, input)
	if err != nil {
		users := setSelected(h.userOptions(ctx), data.AssignedTo)
		contacts := setSelected(h.contactOptionsWithNone(ctx), data.ContactID)
		deals := setSelected(h.dealOptions(ctx), data.DealID)
		pages.TaskFormPage(h.pc(r, "New Task", "/tasks"), data, validationErrors(err), users, contacts, deals).Render(ctx, w)
		return
	}

	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *PageHandler) TaskEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	task, err := h.tasks.GetByID(ctx, h.orgID, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	dueDate := ""
	if task.DueDate != nil {
		dueDate = task.DueDate.Format("2006-01-02")
	}

	data := view.TaskFormData{
		ID:         task.ID.String(),
		Title:      task.Title,
		DueDate:    dueDate,
		ContactID:  uuidPtrToString(task.ContactID),
		DealID:     uuidPtrToString(task.DealID),
		AssignedTo: task.AssignedTo.String(),
	}

	users := setSelected(h.userOptions(ctx), data.AssignedTo)
	contacts := setSelected(h.contactOptionsWithNone(ctx), data.ContactID)
	deals := setSelected(h.dealOptions(ctx), data.DealID)
	pages.TaskFormPage(h.pc(r, "Edit Task", "/tasks"), data, nil, users, contacts, deals).Render(ctx, w)
}

func (h *PageHandler) TaskUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()

	data := view.TaskFormData{
		ID:         id.String(),
		Title:      r.FormValue("title"),
		DueDate:    r.FormValue("due_date"),
		ContactID:  r.FormValue("contact_id"),
		DealID:     r.FormValue("deal_id"),
		AssignedTo: r.FormValue("assigned_to"),
	}

	input := service.TaskUpdateInput{
		Title:      strPtr(data.Title),
		ContactID:  parseOptionalUUID(data.ContactID),
		DealID:     parseOptionalUUID(data.DealID),
		AssignedTo: parseOptionalUUID(data.AssignedTo),
	}
	if data.DueDate != "" {
		t, err := time.Parse("2006-01-02", data.DueDate)
		if err == nil {
			input.DueDate = &t
		}
	}

	_, err = h.tasks.Update(ctx, h.orgID, id, input)
	if err != nil {
		log.Printf("task update: %v", err)
		users := setSelected(h.userOptions(ctx), data.AssignedTo)
		contacts := setSelected(h.contactOptionsWithNone(ctx), data.ContactID)
		deals := setSelected(h.dealOptions(ctx), data.DealID)
		pages.TaskFormPage(h.pc(r, "Edit Task", "/tasks"), data, validationErrors(err), users, contacts, deals).Render(ctx, w)
		return
	}

	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

// contactOptionsWithNone adds a "None" option at the top for optional contact fields.
func (h *PageHandler) contactOptionsWithNone(ctx context.Context) []view.SelectOption {
	opts := h.contactOptions(ctx)
	return append([]view.SelectOption{{Value: "", Label: "— None —"}}, opts...)
}
