package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dukerupert/cairnpost/internal/config"
	"github.com/dukerupert/cairnpost/internal/database"
	"github.com/dukerupert/cairnpost/internal/handler"
	"github.com/dukerupert/cairnpost/internal/repository"
	"github.com/dukerupert/cairnpost/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()

	// Repositories
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)
	contactRepo := repository.NewContactRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	dealRepo := repository.NewDealRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Services
	orgSvc := service.NewOrgService(orgRepo)
	userSvc := service.NewUserService(userRepo)
	companySvc := service.NewCompanyService(companyRepo)
	contactSvc := service.NewContactService(contactRepo, companyRepo)
	activitySvc := service.NewActivityService(activityRepo, contactRepo)
	dealSvc := service.NewDealService(dealRepo, contactRepo)
	taskSvc := service.NewTaskService(taskRepo, userRepo)

	// Resolve org at startup (single-org v1)
	org, err := orgRepo.GetBySlug(context.Background(), cfg.OrgSlug)
	if err != nil {
		log.Fatalf("resolving org slug %q: %v", cfg.OrgSlug, err)
	}
	log.Printf("Resolved org: %s (id=%s)", org.Name, org.ID)

	secureCookie := cfg.Environment != "development"

	// API handlers
	orgH := handler.NewOrgHandler(orgSvc)
	userH := handler.NewUserHandler(userSvc)
	companyH := handler.NewCompanyHandler(companySvc)
	contactH := handler.NewContactHandler(contactSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	dealH := handler.NewDealHandler(dealSvc)
	taskH := handler.NewTaskHandler(taskSvc)

	// Auth handler
	authH := handler.NewAuthHandler(org.ID, userSvc, sessionRepo, secureCookie)

	// Page handler (HTML routes — protected by auth middleware)
	pageH := handler.NewPageHandler(
		org.ID,
		contactSvc, companySvc, dealSvc, taskSvc, activitySvc,
		contactRepo, userRepo, companyRepo,
	)

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Static files (public)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Auth routes (public)
	authH.RegisterRoutes(mux)

	// Protected HTML pages
	protectedMux := http.NewServeMux()
	pageH.RegisterRoutes(protectedMux)
	authMiddleware := handler.RequireAuth(sessionRepo, userRepo, org.ID, secureCookie)
	mux.Handle("/", authMiddleware(protectedMux))

	// Org routes (API, not behind auth)
	orgH.RegisterRoutes(mux)

	// Org-scoped API routes (not behind auth for v1)
	prefix := "/api/v1/{org}"
	companyH.RegisterRoutes(mux, prefix)
	userH.RegisterRoutes(mux, prefix)
	contactH.RegisterRoutes(mux, prefix)
	activityH.RegisterRoutes(mux, prefix)
	dealH.RegisterRoutes(mux, prefix)
	taskH.RegisterRoutes(mux, prefix)

	// Wrap with org resolver middleware (for API routes)
	wrapped := handler.OrgResolver(orgRepo)(mux)

	log.Printf("CairnPost starting on :%s (env=%s)", cfg.Port, cfg.Environment)
	if err := http.ListenAndServe(":"+cfg.Port, wrapped); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
