package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dukerupert/cairnpost/internal/auth"
	"github.com/dukerupert/cairnpost/internal/database"
	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func main() {
	name := flag.String("name", "", "User name (required for admin mode)")
	email := flag.String("email", "", "User email (required)")
	password := flag.String("password", "", "User password (required)")
	orgSlug := flag.String("org", "", "Org slug (defaults to ORG_SLUG env var)")
	demo := flag.Bool("demo", false, "Seed full demo dataset")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  seed --name NAME --email EMAIL --password PASSWORD   # create admin user")
		fmt.Fprintln(os.Stderr, "  seed --demo --email EMAIL --password PASSWORD         # seed full demo dataset")
		os.Exit(1)
	}

	if !*demo && *name == "" {
		fmt.Fprintln(os.Stderr, "--name is required in admin mode (or use --demo)")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	slug := *orgSlug
	if slug == "" {
		slug = os.Getenv("ORG_SLUG")
	}
	if slug == "" {
		slug = "demo"
	}

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if *demo {
		seedDemo(ctx, db, slug, *email, *password)
	} else {
		seedAdmin(ctx, db, slug, *name, *email, *password)
	}
}

func seedAdmin(ctx context.Context, db *sqlx.DB, slug, name, email, password string) {
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)

	org, err := orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		log.Fatalf("org %q not found: %v", slug, err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("hashing password: %v", err)
	}

	existingUser, err := userRepo.GetByEmail(ctx, org.ID, email)
	if err == nil {
		if err := userRepo.SetPasswordHash(ctx, org.ID, existingUser.ID, hash); err != nil {
			log.Fatalf("updating password: %v", err)
		}
		fmt.Printf("Updated password for %s (%s)\n", existingUser.Name, existingUser.Email)
		return
	}

	u := model.User{
		OrgID:        org.ID,
		Name:         name,
		Email:        email,
		Role:         model.RoleAdmin,
		PasswordHash: &hash,
	}
	if err := userRepo.Create(ctx, &u); err != nil {
		log.Fatalf("creating user: %v", err)
	}
	fmt.Printf("Created admin user %s (%s) in org %s\n", u.Name, u.Email, org.Name)
}

func seedDemo(ctx context.Context, db *sqlx.DB, slug, email, password string) {
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	contactRepo := repository.NewContactRepository(db)
	dealRepo := repository.NewDealRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	// 1. Create or find org
	org, err := orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		org = model.Org{Name: "Firefly Software", Slug: slug}
		if err := orgRepo.Create(ctx, &org); err != nil {
			log.Fatalf("creating org: %v", err)
		}
		fmt.Printf("Created org: %s (slug=%s)\n", org.Name, org.Slug)
	} else {
		fmt.Printf("Using existing org: %s (slug=%s)\n", org.Name, org.Slug)
	}

	// 2. Create or find admin user
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("hashing password: %v", err)
	}

	user, err := userRepo.GetByEmail(ctx, org.ID, email)
	if err != nil {
		user = model.User{
			OrgID:        org.ID,
			Name:         "Admin",
			Email:        email,
			Role:         model.RoleAdmin,
			PasswordHash: &hash,
		}
		if err := userRepo.Create(ctx, &user); err != nil {
			log.Fatalf("creating user: %v", err)
		}
		fmt.Printf("Created admin user: %s\n", user.Email)
	} else {
		userRepo.SetPasswordHash(ctx, org.ID, user.ID, hash)
		fmt.Printf("Using existing user: %s\n", user.Email)
	}

	// 3. Companies
	companies := []model.Company{
		{OrgID: org.ID, Name: "Mountain View Builders", Address: "1200 Prospect Ave, Helena, MT", Website: "mountainviewbuilders.com"},
		{OrgID: org.ID, Name: "Helena Plumbing & Heating", Address: "456 Last Chance Gulch, Helena, MT", Website: "helenaplumbing.com"},
		{OrgID: org.ID, Name: "Big Sky Electric", Address: "789 Montana Ave, Helena, MT", Website: "bigskyelectric.com"},
		{OrgID: org.ID, Name: "Glacier Landscaping", Address: "321 Rodney St, Helena, MT"},
		{OrgID: org.ID, Name: "Bridger Roofing", Address: "555 Euclid Ave, Helena, MT", Website: "bridgerroofing.com"},
	}
	for i := range companies {
		if err := companyRepo.Create(ctx, &companies[i]); err != nil {
			log.Fatalf("creating company %s: %v", companies[i].Name, err)
		}
	}
	fmt.Printf("Created %d companies\n", len(companies))

	// 4. Contacts
	contacts := []model.Contact{
		{OrgID: org.ID, FirstName: "Mike", LastName: "Anderson", Email: "mike@mountainviewbuilders.com", Phone: "(406) 555-0101", Tags: pq.StringArray{"contractor", "vip"}, CompanyID: &companies[0].ID},
		{OrgID: org.ID, FirstName: "Sarah", LastName: "Chen", Email: "sarah@mountainviewbuilders.com", Phone: "(406) 555-0102", Tags: pq.StringArray{"contractor"}, CompanyID: &companies[0].ID},
		{OrgID: org.ID, FirstName: "Tom", LastName: "Baker", Email: "tom@helenaplumbing.com", Phone: "(406) 555-0201", Tags: pq.StringArray{"contractor", "referral"}, CompanyID: &companies[1].ID},
		{OrgID: org.ID, FirstName: "Lisa", LastName: "Nguyen", Email: "lisa@bigskyelectric.com", Phone: "(406) 555-0301", Tags: pq.StringArray{"contractor"}, CompanyID: &companies[2].ID},
		{OrgID: org.ID, FirstName: "Jake", LastName: "Williams", Email: "jake@glacierlands.com", Phone: "(406) 555-0401", Tags: pq.StringArray{"lead"}, CompanyID: &companies[3].ID},
		{OrgID: org.ID, FirstName: "Emily", LastName: "Martinez", Email: "emily@bridgerroofing.com", Phone: "(406) 555-0501", Tags: pq.StringArray{"contractor", "vip"}, CompanyID: &companies[4].ID},
		{OrgID: org.ID, FirstName: "David", LastName: "Thompson", Email: "david.t@gmail.com", Phone: "(406) 555-0601", Tags: pq.StringArray{"homeowner", "referral"}},
		{OrgID: org.ID, FirstName: "Rachel", LastName: "Kim", Email: "rachel.kim@outlook.com", Phone: "(406) 555-0701", Tags: pq.StringArray{"homeowner"}},
		{OrgID: org.ID, FirstName: "Chris", LastName: "Davis", Email: "chris@davisconsulting.com", Phone: "(406) 555-0801", Tags: pq.StringArray{"consultant"}},
		{OrgID: org.ID, FirstName: "Amy", LastName: "Johnson", Email: "amy.j@yahoo.com", Phone: "(406) 555-0901", Tags: pq.StringArray{"lead", "cold"}},
	}
	for i := range contacts {
		if err := contactRepo.Create(ctx, &contacts[i]); err != nil {
			log.Fatalf("creating contact %s %s: %v", contacts[i].FirstName, contacts[i].LastName, err)
		}
	}
	fmt.Printf("Created %d contacts\n", len(contacts))

	// 5. Deals
	now := time.Now()
	wonAt := now.Add(-10 * 24 * time.Hour)
	lostAt := now.Add(-5 * 24 * time.Hour)
	deals := []model.Deal{
		{OrgID: org.ID, Title: "Kitchen remodel estimate", Stage: "New Lead", Value: decimal.NewFromInt(15000), ContactID: contacts[0].ID, CompanyID: &companies[0].ID},
		{OrgID: org.ID, Title: "Office plumbing upgrade", Stage: "Estimate Sent", Value: decimal.NewFromInt(8500), ContactID: contacts[2].ID, CompanyID: &companies[1].ID},
		{OrgID: org.ID, Title: "Warehouse electrical panel", Stage: "Estimate Sent", Value: decimal.NewFromInt(22000), ContactID: contacts[3].ID, CompanyID: &companies[2].ID},
		{OrgID: org.ID, Title: "Backyard landscape design", Stage: "Follow-up", Value: decimal.NewFromInt(6000), ContactID: contacts[4].ID, CompanyID: &companies[3].ID},
		{OrgID: org.ID, Title: "Roof replacement", Stage: "Won", Value: decimal.NewFromInt(45000), ContactID: contacts[5].ID, CompanyID: &companies[4].ID, ClosedAt: &wonAt},
		{OrgID: org.ID, Title: "Deck addition", Stage: "New Lead", Value: decimal.NewFromInt(12000), ContactID: contacts[6].ID},
		{OrgID: org.ID, Title: "Bathroom renovation", Stage: "Lost", Value: decimal.NewFromInt(9500), ContactID: contacts[7].ID, ClosedAt: &lostAt},
		{OrgID: org.ID, Title: "IT consulting retainer", Stage: "Follow-up", Value: decimal.NewFromInt(3000), ContactID: contacts[8].ID},
	}
	for i := range deals {
		if err := dealRepo.Create(ctx, &deals[i]); err != nil {
			log.Fatalf("creating deal %s: %v", deals[i].Title, err)
		}
	}
	fmt.Printf("Created %d deals\n", len(deals))

	// 6. Tasks
	yesterday := now.Add(-24 * time.Hour)
	today := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, now.Location())
	tomorrow := now.Add(24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)
	twoDaysAgo := now.Add(-2 * 24 * time.Hour)

	tasks := []model.Task{
		{OrgID: org.ID, Title: "Follow up on kitchen remodel estimate", DueDate: &yesterday, ContactID: &contacts[0].ID, DealID: &deals[0].ID, AssignedTo: user.ID},
		{OrgID: org.ID, Title: "Send plumbing proposal to Tom", DueDate: &today, ContactID: &contacts[2].ID, DealID: &deals[1].ID, AssignedTo: user.ID},
		{OrgID: org.ID, Title: "Schedule site visit for electrical panel", DueDate: &tomorrow, ContactID: &contacts[3].ID, DealID: &deals[2].ID, AssignedTo: user.ID},
		{OrgID: org.ID, Title: "Call Jake about landscape timeline", DueDate: &nextWeek, ContactID: &contacts[4].ID, DealID: &deals[3].ID, AssignedTo: user.ID},
		{OrgID: org.ID, Title: "Send invoice for roof replacement", DueDate: &twoDaysAgo, Done: true, ContactID: &contacts[5].ID, DealID: &deals[4].ID, AssignedTo: user.ID},
		{OrgID: org.ID, Title: "Review David's deck plans", DueDate: &tomorrow, ContactID: &contacts[6].ID, DealID: &deals[5].ID, AssignedTo: user.ID},
	}
	for i := range tasks {
		if err := taskRepo.Create(ctx, &tasks[i]); err != nil {
			log.Fatalf("creating task %s: %v", tasks[i].Title, err)
		}
	}
	fmt.Printf("Created %d tasks\n", len(tasks))

	// 7. Activities
	activities := []model.Activity{
		{OrgID: org.ID, Type: model.ActivityCall, Body: "Discussed kitchen remodel scope. Client wants granite countertops and new cabinets.", ContactID: contacts[0].ID, DealID: &deals[0].ID, UserID: user.ID, OccurredAt: now.Add(-3 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityEmail, Body: "Sent initial estimate for kitchen work. Waiting for approval.", ContactID: contacts[0].ID, DealID: &deals[0].ID, UserID: user.ID, OccurredAt: now.Add(-2 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityNote, Body: "Tom mentioned they also need water heater replacement. Follow up next week.", ContactID: contacts[2].ID, DealID: &deals[1].ID, UserID: user.ID, OccurredAt: now.Add(-4 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivitySiteVisit, Body: "Inspected warehouse electrical panel. Needs full upgrade to 400A service.", ContactID: contacts[3].ID, DealID: &deals[2].ID, UserID: user.ID, OccurredAt: now.Add(-5 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityEmail, Body: "Sent detailed estimate with timeline for electrical upgrade.", ContactID: contacts[3].ID, DealID: &deals[2].ID, UserID: user.ID, OccurredAt: now.Add(-3 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityCall, Body: "Jake wants to start landscaping in spring. Will finalize design in February.", ContactID: contacts[4].ID, DealID: &deals[3].ID, UserID: user.ID, OccurredAt: now.Add(-7 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityNote, Body: "Roof replacement completed successfully. Emily very happy with the work.", ContactID: contacts[5].ID, DealID: &deals[4].ID, UserID: user.ID, OccurredAt: now.Add(-10 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityCall, Body: "David called about adding a covered deck. Wants a quote by end of week.", ContactID: contacts[6].ID, DealID: &deals[5].ID, UserID: user.ID, OccurredAt: now.Add(-1 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivitySMS, Body: "Rachel decided to go with another contractor. Lost deal.", ContactID: contacts[7].ID, DealID: &deals[6].ID, UserID: user.ID, OccurredAt: now.Add(-5 * 24 * time.Hour)},
		{OrgID: org.ID, Type: model.ActivityEmail, Body: "Sent consulting proposal for monthly IT retainer.", ContactID: contacts[8].ID, DealID: &deals[7].ID, UserID: user.ID, OccurredAt: now.Add(-2 * 24 * time.Hour)},
	}
	for i := range activities {
		if err := activityRepo.Create(ctx, &activities[i]); err != nil {
			log.Fatalf("creating activity: %v", err)
		}
	}
	fmt.Printf("Created %d activities\n", len(activities))

	fmt.Println("\nDemo data seeded successfully!")
	fmt.Printf("Login with: %s / %s\n", email, password)
}

// ptr returns a pointer to a uuid — helper for nullable FK fields
func ptr(id uuid.UUID) *uuid.UUID {
	return &id
}
