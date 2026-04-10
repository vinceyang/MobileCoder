package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/config"
	"github.com/mobile-coder/cloud/internal/db"
	"github.com/mobile-coder/cloud/internal/handler"
	"github.com/mobile-coder/cloud/internal/service"
	"github.com/mobile-coder/cloud/internal/ws"
)

// Static file handler for Capacitor app
func staticHandler(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default to index.html for SPA routing
		requestedPath := r.URL.Path
		if requestedPath == "/" || requestedPath == "" {
			requestedPath = "/index.html"
		}

		filePath := filepath.Join(dir, requestedPath)

		// Security: prevent directory traversal
		absDir, _ := filepath.Abs(dir)
		absPath, _ := filepath.Abs(filePath)
		if !filepath.HasPrefix(absPath, absDir) {
			http.NotFound(w, r)
			return
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// For SPA, serve index.html for non-file paths
			indexPath := filepath.Join(dir, "index.html")
			http.ServeFile(w, r, indexPath)
			return
		}

		http.ServeFile(w, r, filePath)
	})
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.Load()

	// Initialize Supabase DB via REST API
	database, err := db.InitDB(&db.Config{
		Host:       cfg.DBHost,
		Port:       cfg.DBPort,
		User:       cfg.DBUser,
		Password:   cfg.DBPassword,
		DBName:     cfg.DBName,
		APIKey:     cfg.SupabaseAPIKey,
		ProjectURL: cfg.SupabaseProjectURL,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize services
	deviceService := service.NewDeviceService(database)
	hub := ws.NewHub()
	taskService := service.NewTaskService(deviceService, hub)
	notificationService := service.NewNotificationService(database)
	tokenManager := cloudauth.NewManager(cfg.JWTSecret, 24*time.Hour)

	// Start WebSocket hub
	go hub.Run()

	// Initialize handlers
	deviceHandler := handler.NewDeviceHandler(deviceService, tokenManager)
	taskHandler := handler.NewTaskHandler(taskService, tokenManager)
	notificationHandler := handler.NewNotificationHandler(notificationService, tokenManager, taskService)
	authService := service.NewAuthService(database)
	authHandler := handler.NewAuthHandler(authService, tokenManager)
	wsHandler := handler.NewWSHubHandler(hub, deviceService, tokenManager)

	// Routes
	mux := http.NewServeMux()

	// Static files for Capacitor app (mobile)
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir != "" {
		mux.Handle("/", staticHandler(staticDir))
		log.Printf("Serving static files from: %s", staticDir)
	}

	// Auth routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// Device routes
	mux.HandleFunc("/api/device/register", deviceHandler.Register)
	mux.HandleFunc("/api/device/bind", deviceHandler.BindDevice)
	mux.HandleFunc("/api/device/bind-agent", deviceHandler.BindAgent)
	mux.HandleFunc("/api/device/list", deviceHandler.ListDevices)
	mux.HandleFunc("/api/device/check", deviceHandler.CheckDevice)
	mux.HandleFunc("/api/device/update", deviceHandler.UpdateDevice)
	mux.HandleFunc("/api/device/delete", deviceHandler.DeleteDevice)
	mux.HandleFunc("/api/devices", deviceHandler.GetUserDevices)
	mux.HandleFunc("/api/tasks", taskHandler.GetTasks)
	mux.HandleFunc("/api/tasks/detail", taskHandler.GetTask)
	mux.HandleFunc("/api/notifications", notificationHandler.ListNotifications)
	mux.HandleFunc("/api/notifications/read", notificationHandler.MarkNotificationRead)
	mux.HandleFunc("/api/notifications/read-all", notificationHandler.MarkAllNotificationsRead)
	mux.HandleFunc("/api/devices/sessions", deviceHandler.GetDeviceSessions)
	mux.HandleFunc("/api/sessions", deviceHandler.CreateSession)
	mux.HandleFunc("/api/sessions/delete", deviceHandler.DeleteSession)

	// WebSocket
	mux.HandleFunc("/ws", wsHandler.HandleConnection)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Wrap with CORS
	handler := corsMiddleware(mux)

	log.Printf("Cloud server starting on port %s (no auth mode)", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
