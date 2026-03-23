package main

import (
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

	// Services
	orgSvc := service.NewOrgService(orgRepo)
	userSvc := service.NewUserService(userRepo)
	companySvc := service.NewCompanyService(companyRepo)
	contactSvc := service.NewContactService(contactRepo, companyRepo)
	activitySvc := service.NewActivityService(activityRepo, contactRepo)
	dealSvc := service.NewDealService(dealRepo, contactRepo)
	taskSvc := service.NewTaskService(taskRepo, userRepo)

	// Handlers
	orgH := handler.NewOrgHandler(orgSvc)
	userH := handler.NewUserHandler(userSvc)
	companyH := handler.NewCompanyHandler(companySvc)
	contactH := handler.NewContactHandler(contactSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	dealH := handler.NewDealHandler(dealSvc)
	taskH := handler.NewTaskHandler(taskSvc)

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Org routes (not behind org middleware)
	orgH.RegisterRoutes(mux)

	// Org-scoped routes
	prefix := "/api/v1/{org}"
	companyH.RegisterRoutes(mux, prefix)
	userH.RegisterRoutes(mux, prefix)
	contactH.RegisterRoutes(mux, prefix)
	activityH.RegisterRoutes(mux, prefix)
	dealH.RegisterRoutes(mux, prefix)
	taskH.RegisterRoutes(mux, prefix)

	// Wrap with org resolver middleware
	wrapped := handler.OrgResolver(orgRepo)(mux)

	log.Printf("CairnPost starting on :%s (env=%s)", cfg.Port, cfg.Environment)
	if err := http.ListenAndServe(":"+cfg.Port, wrapped); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
