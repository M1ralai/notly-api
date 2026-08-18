package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/M1ralai/notly-api/internal/infrastructure"
	"github.com/M1ralai/notly-api/internal/infrastructure/database"
	"github.com/M1ralai/notly-api/internal/infrastructure/email"
	"github.com/M1ralai/notly-api/internal/infrastructure/encryption"
	"github.com/M1ralai/notly-api/internal/infrastructure/google"
	"github.com/M1ralai/notly-api/internal/infrastructure/jobs"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	"github.com/M1ralai/notly-api/internal/infrastructure/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	authHttp "github.com/M1ralai/notly-api/internal/modules/auth/http"
	authRepo "github.com/M1ralai/notly-api/internal/modules/auth/repository"
	authService "github.com/M1ralai/notly-api/internal/modules/auth/service"
	courseHttp "github.com/M1ralai/notly-api/internal/modules/course/http"
	courseRepo "github.com/M1ralai/notly-api/internal/modules/course/repository"
	courseService "github.com/M1ralai/notly-api/internal/modules/course/service"
	eventHttp "github.com/M1ralai/notly-api/internal/modules/event/http"
	eventRepo "github.com/M1ralai/notly-api/internal/modules/event/repository"
	eventService "github.com/M1ralai/notly-api/internal/modules/event/service"
	goalHttp "github.com/M1ralai/notly-api/internal/modules/goal/http"
	goalRepo "github.com/M1ralai/notly-api/internal/modules/goal/repository"
	goalService "github.com/M1ralai/notly-api/internal/modules/goal/service"
	habitHttp "github.com/M1ralai/notly-api/internal/modules/habit/http"
	habitRepo "github.com/M1ralai/notly-api/internal/modules/habit/repository"
	habitService "github.com/M1ralai/notly-api/internal/modules/habit/service"
	lifeareaHttp "github.com/M1ralai/notly-api/internal/modules/lifearea/http"
	lifeareaRepo "github.com/M1ralai/notly-api/internal/modules/lifearea/repository"
	lifeareaService "github.com/M1ralai/notly-api/internal/modules/lifearea/service"
	semesterHttp "github.com/M1ralai/notly-api/internal/modules/semester/http"
	semesterRepo "github.com/M1ralai/notly-api/internal/modules/semester/repository"
	semesterService "github.com/M1ralai/notly-api/internal/modules/semester/service"
	taskHttp "github.com/M1ralai/notly-api/internal/modules/task/http"
	taskRepo "github.com/M1ralai/notly-api/internal/modules/task/repository"
	taskService "github.com/M1ralai/notly-api/internal/modules/task/service"
	userHttp "github.com/M1ralai/notly-api/internal/modules/user/http"
	userRepo "github.com/M1ralai/notly-api/internal/modules/user/repository"
	userService "github.com/M1ralai/notly-api/internal/modules/user/service"

	calendarHttp "github.com/M1ralai/notly-api/internal/modules/calendar/http"
	calendarRepo "github.com/M1ralai/notly-api/internal/modules/calendar/repository"
	calendarService "github.com/M1ralai/notly-api/internal/modules/calendar/service"

	scheduleHttp "github.com/M1ralai/notly-api/internal/modules/schedule/http"
	scheduleRepo "github.com/M1ralai/notly-api/internal/modules/schedule/repository"
	scheduleService "github.com/M1ralai/notly-api/internal/modules/schedule/service"

	pomodoroHttp "github.com/M1ralai/notly-api/internal/modules/pomodoro/http"
	pomodoroRepo "github.com/M1ralai/notly-api/internal/modules/pomodoro/repository"
	pomodoroService "github.com/M1ralai/notly-api/internal/modules/pomodoro/service"

	"github.com/M1ralai/notly-api/internal/infrastructure/websocket"

	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"

	dashboardAdapters "github.com/M1ralai/notly-api/internal/adapters/dashboard"
	dashboardHttp "github.com/M1ralai/notly-api/internal/modules/dashboard/http"
	dashboardService "github.com/M1ralai/notly-api/internal/modules/dashboard/service"

	subscriptionHttp "github.com/M1ralai/notly-api/internal/modules/subscription/http"
	subscriptionRepo "github.com/M1ralai/notly-api/internal/modules/subscription/repository"
	subscriptionService "github.com/M1ralai/notly-api/internal/modules/subscription/service"
	syncHttp "github.com/M1ralai/notly-api/internal/modules/sync/http"
	syncService "github.com/M1ralai/notly-api/internal/modules/sync/service"

	noteHttp "github.com/M1ralai/notly-api/internal/modules/note/http"
	noteRepo "github.com/M1ralai/notly-api/internal/modules/note/repository"
	noteService "github.com/M1ralai/notly-api/internal/modules/note/service"

	"github.com/M1ralai/notly-api/internal/adapters/storage"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	httpServer *http.Server
	db         *sqlx.DB
	logger     *logger.ZapLogger
}

// Sadece localhost üzerinden erişime izin veren middleware
func LocalhostOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		ip = strings.Trim(ip, "[]")

		// Sadece localhost erişimi
		if ip != "127.0.0.1" && ip != "::1" && ip != "localhost" {
			http.Error(w, "Forbidden - Sadece localhost erişimine izin verilmiştir (IP: "+ip+")", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewServer(db *sqlx.DB, zapLogger *logger.ZapLogger) *Server {
	// WebSocket Hub
	wsHub := websocket.NewHub(zapLogger)
	go wsHub.Run()
	wsHandler := websocket.NewHandler(wsHub, zapLogger)

	// Broadcaster for real-time notifications
	broadcaster := notifService.NewBroadcaster(wsHub, zapLogger)

	// Distributed lock for jobs
	jobLock := jobs.NewDistributedLock(db)

	// Job Pool for async task processing
	jobPool := jobs.NewWorkerPool(5, 100, zapLogger, nil, jobLock)
	jobPool.Start()

	// User module
	userRepository := userRepo.NewPostgresRepository(db)

	// Subscription module
	subscriptionRepository := subscriptionRepo.NewPostgresRepository(db)
	subscriptionSvc := subscriptionService.NewSubscriptionService(subscriptionRepository)
	subscriptionHandler := subscriptionHttp.NewHandler(subscriptionSvc)

	userSvc := userService.NewUserService(userRepository, zapLogger, subscriptionSvc)
	userHandler := userHttp.NewHandler(userSvc)

	// Auth module
	dbWrapper := &database.Database{Conn: db}
	emailService := email.NewResendEmailService()
	turnstileValidator := infrastructure.NewTurnstileValidator()
	refreshTokenRepo := authRepo.NewPostgresRepository(db)
	authSvc := authService.NewAuthService(userRepository, refreshTokenRepo, zapLogger, dbWrapper, emailService, jobPool, turnstileValidator, subscriptionSvc)
	authHandler := authHttp.NewHandler(authSvc)

	// LifeArea module
	lifeareaRepository := lifeareaRepo.NewPostgresRepository(db)
	lifeareaSvc := lifeareaService.NewLifeAreaService(lifeareaRepository, zapLogger, broadcaster)
	lifeareaHandler := lifeareaHttp.NewHandler(lifeareaSvc)

	// Semester module (before course module, as course depends on it)
	semesterRepository := semesterRepo.NewPostgresRepository(db)
	semesterSvc := semesterService.NewSemesterService(semesterRepository, zapLogger, broadcaster)
	semesterHandler := semesterHttp.NewHandler(semesterSvc)

	// Calendar module (created early for injection into other handlers)
	integrationRepository := calendarRepo.NewCalendarIntegrationRepository(db)
	syncQueueRepository := calendarRepo.NewSyncQueueRepository(db)
	eventMappingRepository := calendarRepo.NewEventMappingRepository(db)

	// Google Calendar client (optional - gracefully handle if not configured)
	var googleClient google.CalendarClient
	googleClient, _ = google.NewCalendarClient() // Ignore error if not configured

	// Encryption for tokens (optional - gracefully handle if not configured)
	var encryptor *encryption.Encryptor
	encryptor, _ = encryption.NewEncryptor() // Ignore error if not configured

	calendarSvc := calendarService.NewCalendarService(
		integrationRepository,
		syncQueueRepository,
		eventMappingRepository,
		googleClient,
		encryptor,
		zapLogger,
		broadcaster,
	)
	calendarHandler := calendarHttp.NewHandler(calendarSvc)

	// Shared object storage adapter for Pro file features
	minioAdapter, minioErr := storage.NewMinIOAdapter()
	if minioErr != nil {
		log.Printf("⚠ MinIO not available (%v) – file uploads disabled", minioErr)
	}
	var storageProvider storage.StorageProvider
	if minioAdapter != nil {
		if err := minioAdapter.EnsureBucket(); err != nil {
			log.Printf("⚠ MinIO bucket init failed: %v", err)
		} else {
			storageProvider = minioAdapter
			log.Println("✓ MinIO storage adapter ready")
		}
	}

	// Course module (after calendar for sync injection)
	courseRepository := courseRepo.NewPostgresRepository(db)
	courseSvc := courseService.NewCourseService(courseRepository, semesterRepository, calendarSvc, zapLogger, broadcaster, storageProvider, subscriptionSvc, userRepository)
	courseHandler := courseHttp.NewHandler(courseSvc)

	// Task module (after calendar for sync injection)
	taskRepository := taskRepo.NewPostgresRepository(db)
	taskSvc := taskService.NewTaskService(taskRepository, zapLogger, broadcaster, calendarSvc)
	taskHandler := taskHttp.NewHandler(taskSvc, jobPool, taskRepository, broadcaster, calendarSvc, zapLogger)

	// Habit module (after calendar for sync injection)
	habitRepository := habitRepo.NewPostgresRepository(db)
	habitSvc := habitService.NewHabitService(habitRepository, zapLogger, broadcaster)
	habitHandler := habitHttp.NewHandler(habitSvc, habitRepository, broadcaster, calendarSvc, zapLogger)

	// Goal module
	goalRepository := goalRepo.NewPostgresRepository(db)
	goalSvc := goalService.NewGoalService(goalRepository, zapLogger, broadcaster)
	goalHandler := goalHttp.NewHandler(goalSvc)

	// Event module
	eventRepository := eventRepo.NewPostgresRepository(db)
	eventSvc := eventService.NewEventService(eventRepository, zapLogger, broadcaster)
	eventHandler := eventHttp.NewHandler(eventSvc)

	// Schedule module
	blockedSlotRepository := scheduleRepo.NewBlockedTimeSlotRepository(db)
	scheduleSvc := scheduleService.NewScheduleService(blockedSlotRepository, zapLogger, broadcaster)
	scheduleHandler := scheduleHttp.NewHandler(scheduleSvc)

	// Pomodoro module
	pomodoroRepository := pomodoroRepo.NewPgRepository(db)
	pomodoroSvc := pomodoroService.NewService(pomodoroRepository)
	pomodoroHandler := pomodoroHttp.NewHandler(pomodoroSvc)

	// Note module (shared notes system)
	noteRepository := noteRepo.NewPostgresRepository(db)
	noteSvc := noteService.NewNoteService(noteRepository, storageProvider, subscriptionSvc)
	noteHandler := noteHttp.NewHandler(noteSvc)

	// Sync module (Aggregates Repositories)
	syncSvc := syncService.NewSyncService(taskRepository, habitRepository, lifeareaRepository, goalRepository, courseRepository, eventRepository, semesterRepository, noteRepository, zapLogger)
	syncHandler := syncHttp.NewHandler(syncSvc, broadcaster)

	// Dashboard module (Aggregated)
	taskAdapter := dashboardAdapters.NewTaskAdapter(taskSvc)
	habitAdapter := dashboardAdapters.NewHabitAdapter(habitSvc)
	lifeAreaAdapter := dashboardAdapters.NewLifeAreaAdapter(lifeareaSvc)

	dashboardSvc := dashboardService.NewDashboardService(taskAdapter, habitAdapter, lifeAreaAdapter)
	dashboardHandler := dashboardHttp.NewHandler(dashboardSvc)

	// Analytics module routes were removed

	router := mux.NewRouter()

	// Router içerisine Swagger Endpoint'ini Localhost Korumalı olarak ekleme
	router.PathPrefix("/swagger/").Handler(LocalhostOnlyMiddleware(httpSwagger.WrapHandler))

	// Note module – public shared note route (no JWT)
	noteHandler.RegisterPublicRoutes(router)

	// WebSocket route - MUST be registered BEFORE middleware to avoid ResponseWriter wrapping
	// WebSocket requires the original http.ResponseWriter to hijack the connection
	router.HandleFunc("/ws", wsHandler.HandleConnection).Methods("GET")

	// Apply middleware to all routes EXCEPT WebSocket
	router.Use(middleware.CORSMiddleware) // Must be first — intercepts OPTIONS before mux can 405
	router.Use(middleware.RecoveryMiddleware)
	router.Use(zapLogger.Middleware)
	router.Use(middleware.MetricsMiddleware)
	router.Use(middleware.TimeoutMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip Content-Type header for WebSocket upgrade
			if r.URL.Path != "/ws" {
				w.Header().Set("Content-Type", "application/json")
			}
			next.ServeHTTP(w, r)
		})
	})

	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Register public auth routes directly on the root router (/auth/...)
	authHandler.RegisterRoutes(router)

	// API routes (protected)
	api := router.PathPrefix("/api").Subrouter()

	api.Use(middleware.AuthMiddleware)

	// Register public auth routes on the /api subrouter (/api/auth/...)
	authHandler.RegisterRoutes(api)

	// Register all module routes
	userHandler.RegisterRoutes(api)
	lifeareaHandler.RegisterRoutes(api)
	courseHandler.RegisterRoutes(api)
	semesterHandler.RegisterRoutes(api)
	taskHandler.RegisterRoutes(api)
	habitHandler.RegisterRoutes(api)
	goalHandler.RegisterRoutes(api)
	eventHandler.RegisterRoutes(api)
	calendarHandler.RegisterRoutes(api)
	scheduleHandler.RegisterRoutes(api)
	pomodoroHandler.RegisterRoutes(api)
	dashboardHandler.RegisterRoutes(api)
	subscriptionHandler.RegisterRoutes(api)
	syncHandler.RegisterRoutes(api)
	// Note module – protected routes
	noteHandler.RegisterRoutes(api)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = ":8080"
	}
	if len(port) > 0 && port[0] != ':' {
		port = ":" + port
	}

	httpServer := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		db:         db,
		logger:     zapLogger,
	}
}

func (s *Server) Start() error {
	errChan := make(chan error, 1)

	go func() {
		log.Printf("✓ Server starting... Port: %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):

		return nil
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("✓ Graceful shutdown started...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("database close error: %w", err)
	}

	log.Println("✓ Shutdown completed")
	return nil
}
