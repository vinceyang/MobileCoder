package main

import (
	"log"
	"net/http"

	"github.com/coder/agentapi/cloud/internal/config"
	"github.com/coder/agentapi/cloud/internal/db"
	"github.com/coder/agentapi/cloud/internal/handler"
	"github.com/coder/agentapi/cloud/internal/service"
	"github.com/coder/agentapi/cloud/internal/ws"
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
	wsHandler := handler.NewWSHubHandler(hub)

	// Routes - 简化版（移除用户登录相关）
	mux := http.NewServeMux()

	// Device routes - 简化，移除 token 验证
	mux.HandleFunc("/api/device/register", deviceHandler.Register)
	mux.HandleFunc("/api/device/bind", deviceHandler.BindDevice)
	mux.HandleFunc("/api/device/bind-agent", deviceHandler.BindAgent)
	mux.HandleFunc("/api/device/list", deviceHandler.ListDevices)

	// WebSocket
	mux.HandleFunc("/ws", wsHandler.HandleConnection)

	// Wrap with CORS
	handler := corsMiddleware(mux)

	log.Printf("Cloud server starting on port %s (no auth mode)", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
