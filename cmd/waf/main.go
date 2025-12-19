package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aleh/docode-waf/internal/api"
	"github.com/aleh/docode-waf/internal/config"
	"github.com/aleh/docode-waf/internal/middleware"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/aleh/docode-waf/internal/proxy"
	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using config.yaml or environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		// Try local config
		cfg, err = config.LoadConfig("config.local.yaml")
		if err != nil {
			log.Printf("Warning: Failed to load config file: %v", err)
			// Continue with default config from env vars
			cfg = &config.Config{}
		}
	}

	// Connect to database using helper method
	db, err := sqlx.Connect(cfg.Database.Driver, cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize services
	logService := services.NewLogService(db)
	vhostService := services.NewVHostService(db)

	// Load virtual hosts
	vhosts, err := vhostService.GetAll()
	if err != nil {
		log.Printf("Warning: Failed to load vhosts: %v", err)
		vhosts = []*models.VHost{}
	}

	// Initialize reverse proxy
	reverseProxy := proxy.NewReverseProxy(cfg)
	reverseProxy.LoadVHosts(vhosts)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(
		cfg.WAF.RateLimit.RequestsPerSecond,
		cfg.WAF.RateLimit.Burst,
	)

	httpFloodProtector := middleware.NewHTTPFloodProtector(
		cfg.WAF.HTTPFlood.MaxRequestsPerMinute,
		time.Duration(cfg.WAF.HTTPFlood.BlockDuration)*time.Second,
	)

	ipBlocker := middleware.NewIPBlocker()
	botDetector := middleware.NewBotDetector(true)
	loggingMiddleware := middleware.NewLoggingMiddleware(logService)

	// Create WAF handler chain
	wafHandler := http.Handler(reverseProxy)
	wafHandler = loggingMiddleware.Middleware(wafHandler)

	if cfg.WAF.RateLimit.Enabled {
		wafHandler = rateLimiter.Middleware(wafHandler)
	}

	if cfg.WAF.HTTPFlood.Enabled {
		wafHandler = httpFloodProtector.Middleware(wafHandler)
	}

	if cfg.WAF.AntiBot.Enabled {
		wafHandler = botDetector.Middleware(wafHandler)
	}

	wafHandler = ipBlocker.Middleware(wafHandler)

	// Start WAF server
	wafServer := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      wafHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting WAF server on %s", wafServer.Addr)
		if err := wafServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WAF server error: %v", err)
		}
	}()

	// Initialize API handlers
	vhostHandler := api.NewVHostHandler(db)
	ipGroupHandler := api.NewIPGroupHandler(db)
	dashboardHandler := api.NewDashboardHandler(db)

	// Setup admin API
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	apiV1 := router.Group("/api/v1")
	{
		// Dashboard
		apiV1.GET("/dashboard/stats", dashboardHandler.GetStats)
		apiV1.GET("/dashboard/traffic", dashboardHandler.GetTrafficLogs)

		// Virtual Hosts
		apiV1.GET("/vhosts", vhostHandler.List)
		apiV1.GET("/vhosts/:id", vhostHandler.Get)
		apiV1.POST("/vhosts", vhostHandler.Create)
		apiV1.PUT("/vhosts/:id", vhostHandler.Update)
		apiV1.DELETE("/vhosts/:id", vhostHandler.Delete)

		// IP Groups
		apiV1.GET("/ip-groups", ipGroupHandler.List)
		apiV1.GET("/ip-groups/:id", ipGroupHandler.Get)
		apiV1.POST("/ip-groups", ipGroupHandler.Create)
		apiV1.DELETE("/ip-groups/:id", ipGroupHandler.Delete)
		apiV1.POST("/ip-groups/:id/ips", ipGroupHandler.AddIP)
		apiV1.GET("/ip-groups/:id/ips", ipGroupHandler.ListIPs)
		apiV1.DELETE("/ip-groups/:id/ips/:ipId", ipGroupHandler.RemoveIP)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start admin API server
	adminServer := &http.Server{
		Addr:         cfg.GetAdminAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting Admin API server on %s", adminServer.Addr)
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin API server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wafServer.Shutdown(ctx); err != nil {
		log.Printf("WAF server shutdown error: %v", err)
	}

	if err := adminServer.Shutdown(ctx); err != nil {
		log.Printf("Admin API server shutdown error: %v", err)
	}

	log.Println("Servers stopped")
}
