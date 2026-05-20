package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/hospital_management/backend/internal/database"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/handler"
	"github.com/hospital_management/backend/internal/middleware"
	"github.com/hospital_management/backend/internal/repository"
	"github.com/hospital_management/backend/internal/service"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Database connection
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Repositories
	bedRepo := repository.NewBedRepository(pool)
	bedTypeRepo := repository.NewBedTypeRepository(pool)
	patientRepo := repository.NewPatientRepository(pool)
	admissionRepo := repository.NewAdmissionRepository(pool)
	clinicalLogRepo := repository.NewClinicalLogRepository(pool)
	staffRepo := repository.NewStaffRepository(pool)
	sseTicketRepo := repository.NewSSETicketRepository(pool)
	settingsRepo := repository.NewSettingsRepository(pool)

	// Services
	sseService := service.NewSSEService(pool, sseTicketRepo, settingsRepo)
	bedService := service.NewBedService(bedRepo, patientRepo, admissionRepo)
	bedTypeService := service.NewBedTypeService(bedTypeRepo)
	patientService := service.NewPatientService(patientRepo)
	admissionService := service.NewAdmissionService(pool, admissionRepo, bedRepo, patientRepo, sseService)
	clinicalLogService := service.NewClinicalLogService(pool, clinicalLogRepo, admissionRepo, bedRepo, sseService)
	authService := service.NewAuthService(staffRepo)

	// Background Worker
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go sseService.StartWorker(workerCtx)

	// Handlers
	bedHandler := handler.NewBedHandler(bedService)
	bedTypeHandler := handler.NewBedTypeHandler(bedTypeService)
	patientHandler := handler.NewPatientHandler(patientService)
	admissionHandler := handler.NewAdmissionHandler(admissionService)
	clinicalLogHandler := handler.NewClinicalLogHandler(clinicalLogService, admissionService)
	authHandler := handler.NewAuthHandler(authService)
	staffHandler := handler.NewStaffHandler(authService)
	sseHandler := handler.NewSSEHandler(sseService)
	settingsHandler := handler.NewSettingsHandler(sseService)

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)
	r.Use(chi_middleware.RealIP)

	// CORS
	allowedOriginsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:5173")
	allowedOrigins := strings.Split(allowedOriginsStr, ",")

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, o := range allowedOrigins {
				if origin == o {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Public Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)
		r.Get("/events", sseHandler.StreamEvents)

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)

			r.Post("/auth/sse-ticket", sseHandler.CreateTicket)

			r.Route("/settings", func(r chi.Router) {
				r.Get("/", settingsHandler.GetSettings)
				r.With(middleware.RequireRole(domain.RoleAdmin)).Put("/", settingsHandler.UpdateSettings)
			})

			r.Route("/beds", func(r chi.Router) {
				r.Get("/", bedHandler.GetAll)
				r.Get("/{id}", bedHandler.Get)
				r.Get("/{id}/patient", bedHandler.GetPatient)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole(domain.RoleAdmin))
					r.Post("/", bedHandler.Create)
					r.Put("/{id}", bedHandler.Update)
					r.Delete("/{id}", bedHandler.Delete)
				})
			})

			r.Route("/bed-types", func(r chi.Router) {
				r.Get("/", bedTypeHandler.GetAll)
				r.Get("/{id}", bedTypeHandler.Get)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole(domain.RoleAdmin))
					r.Post("/", bedTypeHandler.Create)
					r.Put("/{id}", bedTypeHandler.Update)
					r.Delete("/{id}", bedTypeHandler.Delete)
				})
			})

			r.Route("/patients", func(r chi.Router) {
				r.Get("/", patientHandler.List)
				r.Post("/", patientHandler.Create)
				r.Get("/search", patientHandler.Search)
				r.Get("/{id}", patientHandler.Get)
				r.Put("/{id}", patientHandler.Update)
			})

			r.Route("/admissions", func(r chi.Router) {
				r.Post("/", admissionHandler.Create)
				r.Get("/{id}", admissionHandler.Get)
				r.Put("/{id}/discharge", admissionHandler.Discharge)
				r.Put("/{id}/event", admissionHandler.RegisterEvent)
				r.Post("/{id}/clinical-logs", clinicalLogHandler.Create)
				r.Get("/{id}/clinical-logs", clinicalLogHandler.List)
			})

			r.Route("/users", func(r chi.Router) {
				r.Use(middleware.RequireRole(domain.RoleAdmin))
				r.Get("/", staffHandler.List)
				r.Post("/", staffHandler.Create)
				r.Put("/{id}/password", staffHandler.ChangePassword)
				r.Put("/{id}/active", staffHandler.SetActive)
			})
		})
	})

	// Server
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}