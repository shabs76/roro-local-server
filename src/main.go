package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/shabs76/roro-local-server/pkg/logger"
	"github.com/shabs76/roro-local-server/pkg/middleware"
	"github.com/shabs76/roro-local-server/routes"
)

func main() {
	// 1. Load Environment Variables
	// It's okay if .env doesn't exist (e.g. inside Docker), we might use real env vars.
	if err := godotenv.Load(); err != nil {
		// Just a warning, not fatal
		slog.Warn("No .env file found, relying on system environment variables")
	}

	// 2. Setup Logger
	// We default to "logs" directory and "server.log" file
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "logs"
	}
	// Initialize our custom logger (writes to file + stdout)
	logger.Setup(logDir, "server.log")

	// 3. Setup Gin
	// Set Gin mode based on env
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() to have control over middleware
	r := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Logger middleware logs each request.
	// Since we set gin.DefaultWriter in logger.Setup, this will log to our file and stdout.
	// r.Use(gin.Logger())
	// Use custom detailed logger
	r.Use(middleware.RequestLogger())

	// 4. Define Routes
	routes.SetupManifestRoutes(r)
	routes.SetupUserRoutes(r)
	routes.SetupSSESyncRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	r.GET("/", func(c *gin.Context) {
		slog.Info("Root handler accessed", "ip", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Roro Local Server",
			"version": "1.0.0",
		})
	})

	// 5. Server Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "4400"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		// Good practice: set timeouts
		ReadTimeout: 300 * time.Second,
		// Removed WriteTimeout because it drops Server-Sent Events (SSE)
		// connections unconditionally if they take longer than the timeout.
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	// 6. Graceful Shutdown
	// Run server in a goroutine so that it doesn't block the graceful shutdown handling below
	go func() {
		slog.Info("Starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
