package main

import (
	"log"
	"net/http"

	"github.com/mobile-coder/cloud/internal/config"
	"github.com/mobile-coder/cloud/internal/db"
	"github.com/mobile-coder/cloud/internal/handler"
	"github.com/mobile-coder/cloud/internal/service"
	"github.com/mobile-coder/cloud/internal/ws"
)

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
		Host:            cfg.DBHost,
		Port:            cfg.DBPort,
		User:            cfg.DBUser,
		Password:        cfg.DBPassword,
		DBName:          cfg.DBName,
		APIKey:          cfg.SupabaseAPIKey,
		ProjectURL:      cfg.SupabaseProjectURL,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize services
	deviceService := service.NewDeviceService(database)
	hub := ws.NewHub()

	// Start WebSocket hub
	go hub.Run()

	// Initialize handlers
	deviceHandler := handler.NewDeviceHandler(deviceService)
	authService := service.NewAuthService(database)
	authHandler := handler.NewAuthHandler(authService)
	wsHandler := handler.NewWSHubHandler(hub, deviceService)

	// Routes
	mux := http.NewServeMux()

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
	mux.HandleFunc("/api/devices/sessions", deviceHandler.GetDeviceSessions)
	mux.HandleFunc("/api/sessions", deviceHandler.CreateSession)

	// WebSocket
	mux.HandleFunc("/ws", wsHandler.HandleConnection)

	// Wrap with CORS
	handler := corsMiddleware(mux)

	log.Printf("Cloud server starting on port %s (no auth mode)", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
