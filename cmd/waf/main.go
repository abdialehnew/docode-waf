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
	"github.com/aleh/docode-waf/internal/constants"
	"github.com/aleh/docode-waf/internal/middleware"
	"github.com/aleh/docode-waf/internal/proxy"
	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func loadEnvironment() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or config.yaml")
	}
}

func initRedis(cfg *config.Config) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return redisClient
}

func initDatabase(cfg *config.Config) *sqlx.DB {
	db, err := sqlx.Connect(cfg.Database.Driver, cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func initServices(db *sqlx.DB) (*services.VHostService, *services.CertificateService, *services.NginxConfigService, *services.AuthService) {
	vhostService := services.NewVHostService(db)
	certService := services.NewCertificateService(db)
	nginxConfigService := services.NewNginxConfigServiceWithDB(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-this-secret-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}
	authService := services.NewAuthService(db, jwtSecret)

	return vhostService, certService, nginxConfigService, authService
}

func setupVHostsAndCerts(vhostService *services.VHostService, certService *services.CertificateService, nginxConfigService *services.NginxConfigService) {
	log.Println("Loading virtual hosts from database...")
	vhosts, err := vhostService.ListVHosts()
	if err != nil {
		log.Printf("Warning: Failed to load virtual hosts: %v", err)
		return
	}

	log.Printf("Loaded %d enabled virtual hosts", len(vhosts))

	// Export SSL certificates for vhosts
	log.Println("Exporting SSL certificates...")
	for _, vhost := range vhosts {
		if vhost.SSLEnabled && vhost.SSLCertificateID != "" {
			cert, err := certService.GetCertificate(vhost.SSLCertificateID)
			if err == nil {
				if err := certService.SaveCertificateFiles(vhost.SSLCertificateID, []byte(cert.CertContent), []byte(cert.KeyContent)); err != nil {
					log.Printf("Warning: Failed to export certificate %s: %v", vhost.SSLCertificateID, err)
				} else {
					log.Printf("Exported certificate for domain: %s", vhost.Domain)
				}
			}
		}
	}

	// Generate nginx configs for all vhosts
	log.Println("Generating nginx configurations...")
	if err := nginxConfigService.RegenerateAllVHostConfigs(vhosts); err != nil {
		log.Printf("Warning: Failed to generate nginx configs: %v", err)
	} else {
		log.Printf("Generated nginx configurations for %d vhosts", len(vhosts))
	}
}

func setupWAFServer(cfg *config.Config, redisClient *redis.Client, db *sqlx.DB, reverseProxyHandler http.Handler) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	wafRouter := gin.New()
	wafRouter.Use(gin.Recovery())

	// Apply WAF middleware
	wafRouter.Use(middleware.RateLimiterMiddleware(redisClient, cfg.WAF.RateLimit.RequestsPerSecond, cfg.WAF.RateLimit.Burst))
	wafRouter.Use(middleware.HTTPFloodProtectionMiddleware(redisClient, cfg.WAF.HTTPFlood.MaxRequestsPerMinute, time.Minute))
	wafRouter.Use(middleware.IPBlockerMiddleware(db))
	wafRouter.Use(middleware.BotDetectorMiddleware())
	wafRouter.Use(middleware.LoggingMiddleware(db))

	// Proxy all requests to the reverse proxy
	wafRouter.NoRoute(gin.WrapH(reverseProxyHandler))

	// Create server
	wafServer := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      wafRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting WAF server on %s", wafServer.Addr)
		if err := wafServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WAF server error: %v", err)
		}
	}()

	return wafServer
}

func setupAPIRoutes(apiV1 *gin.RouterGroup, authService *services.AuthService, authHandler *api.AuthHandler,
	dashboardHandler *api.DashboardHandler, vhostHandler *api.VHostHandler,
	ipGroupHandler *api.IPGroupHandler, certHandler *api.CertificateHandler, settingsHandler *api.SettingsHandler, cfg *config.Config) {

	// Public Auth routes (no authentication required)
	auth := apiV1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/request-reset", authHandler.RequestPasswordReset)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	// Public Settings route (accessible without authentication)
	apiV1.GET("/settings/app", settingsHandler.GetAppSettings)

	// Public Turnstile site key
	apiV1.GET("/turnstile/sitekey", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"site_key": cfg.Turnstile.SiteKey,
			"enabled":  cfg.Turnstile.SiteKey != "" && cfg.Turnstile.SiteKey != "${TURNSTILE_SITE_KEY}",
		})
	})

	// Protected routes (require authentication)
	protected := apiV1.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// Auth profile
		protected.GET("/auth/profile", authHandler.GetProfile)
		protected.POST("/auth/change-password", authHandler.ChangePassword)

		// Dashboard
		protected.GET("/dashboard/stats", dashboardHandler.GetStats)
		protected.GET("/dashboard/traffic", dashboardHandler.GetTrafficLogs)
		protected.GET("/dashboard/attacks", dashboardHandler.GetAttackLogs)
		protected.GET("/dashboard/attack-stats", dashboardHandler.GetAttackStats)
		protected.GET("/dashboard/recent-attacks", dashboardHandler.GetRecentAttacks)
		protected.GET("/dashboard/attacks-by-country", dashboardHandler.GetAttacksByCountry)

		// Virtual Hosts
		protected.GET("/vhosts", vhostHandler.ListVHosts)
		protected.GET(constants.RouteVHostID, vhostHandler.GetVHost)
		protected.POST("/vhosts", vhostHandler.CreateVHost)
		protected.PUT(constants.RouteVHostID, vhostHandler.UpdateVHost)
		protected.DELETE(constants.RouteVHostID, vhostHandler.DeleteVHost)

		// VHost Config Editor
		protected.GET("/vhost-config/:domain", vhostHandler.GetVHostConfig)
		protected.PUT("/vhost-config/:domain", vhostHandler.UpdateVHostConfig)

		// IP Groups
		protected.GET("/ip-groups", ipGroupHandler.ListIPGroups)
		protected.GET("/ip-groups/:id", ipGroupHandler.GetIPGroup)
		protected.POST("/ip-groups", ipGroupHandler.CreateIPGroup)
		protected.DELETE("/ip-groups/:id", ipGroupHandler.DeleteIPGroup)
		protected.POST("/ip-groups/:id/addresses", ipGroupHandler.AddIPAddress)
		protected.DELETE("/ip-groups/:id/addresses/:addressId", ipGroupHandler.DeleteIPAddress)

		// SSL Certificates
		protected.GET("/certificates", certHandler.GetCertificates)
		protected.GET(constants.RouteCertificateID, certHandler.GetCertificate)
		protected.POST("/certificates", certHandler.CreateCertificate)
		protected.POST("/certificates/upload", certHandler.UploadCertificate)
		protected.PUT(constants.RouteCertificateID, certHandler.UpdateCertificate)
		protected.DELETE(constants.RouteCertificateID, certHandler.DeleteCertificate)
		protected.POST("/certificates/update-statuses", certHandler.UpdateCertificateStatuses)

		// Settings (POST only, GET is public)
		protected.POST("/settings/app", settingsHandler.SaveAppSettings)
	}
}

func setupAdminServer(cfg *config.Config, db *sqlx.DB, nginxConfigService *services.NginxConfigService,
	vhostService *services.VHostService, certService *services.CertificateService, authService *services.AuthService) *http.Server {

	// Initialize email service
	emailService := services.NewEmailService(db)

	// Initialize API handlers
	vhostHandler := api.NewVHostHandler(db, nginxConfigService, vhostService, certService)
	ipGroupHandler := api.NewIPGroupHandler(db)
	dashboardHandler := api.NewDashboardHandler(db)
	authHandler := api.NewAuthHandler(authService, emailService, cfg)
	certHandler := api.NewCertificateHandler(certService)
	settingsHandler := api.NewSettingsHandler(db)

	// Setup admin API
	adminRouter := gin.Default()

	// Get allowed origins from config
	allowedOrigins := cfg.GetCORSAllowedOrigins()

	// CORS middleware
	adminRouter.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowOrigin := "*"
		if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			// Allow all origins
			allowOrigin = origin
			if allowOrigin == "" {
				allowOrigin = "*"
			}
		} else {
			// Check if origin is in allowed list
			for _, allowed := range allowedOrigins {
				if allowed == origin || allowed == "*" {
					allowOrigin = origin
					break
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	apiV1 := adminRouter.Group("/api/v1")
	setupAPIRoutes(apiV1, authService, authHandler, dashboardHandler, vhostHandler, ipGroupHandler, certHandler, settingsHandler, cfg)

	// Health check
	adminRouter.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Create server
	adminServer := &http.Server{
		Addr:         cfg.GetAdminAddr(),
		Handler:      adminRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting Admin API server on %s", adminServer.Addr)
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin API server error: %v", err)
		}
	}()

	return adminServer
}

func gracefulShutdown(wafServer, adminServer *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wafServer.Shutdown(ctx); err != nil {
		log.Printf("WAF server forced to shutdown: %v", err)
	}

	if err := adminServer.Shutdown(ctx); err != nil {
		log.Printf("Admin API server forced to shutdown: %v", err)
	}

	log.Println("Servers exited")
}

func main() {
	// Load environment
	loadEnvironment()

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize dependencies
	redisClient := initRedis(cfg)
	db := initDatabase(cfg)
	defer db.Close()

	// Initialize services
	vhostService, certService, nginxConfigService, authService := initServices(db)

	// Initialize reverse proxy
	reverseProxyHandler := proxy.NewReverseProxy(cfg, vhostService)

	// Load virtual hosts and setup certificates
	vhosts, err := vhostService.ListVHosts()
	if err == nil {
		reverseProxyHandler.LoadVHosts(vhosts)
		setupVHostsAndCerts(vhostService, certService, nginxConfigService)
	}

	// Start servers
	wafServer := setupWAFServer(cfg, redisClient, db, reverseProxyHandler)
	adminServer := setupAdminServer(cfg, db, nginxConfigService, vhostService, certService, authService)

	// Wait for shutdown signal
	gracefulShutdown(wafServer, adminServer)
}
