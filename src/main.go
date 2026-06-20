package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/env"
	tenantmw "github.com/hornosg/go-shared/infrastructure/middleware"
	"github.com/hornosg/go-shared/infrastructure/postgres"
	sharedmigrate "github.com/hornosg/go-shared/migrate"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	paymentMethodConfig "payment_method/src/payment_method/infrastructure/config"

	paymentroot "payment_method"
)

func main() {
	// Configuración de la base de datos
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Migraciones versionadas in-app (ADR-001) — fail-fast antes de servir tráfico.
	dbName := env.Get("DB_NAME", "payment_method_db")
	if err := sharedmigrate.RunMigrations(db, paymentroot.MigrationsFS, dbName); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Configuración del router
	router := gin.New()

	// Agregar middlewares básicos
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(tenantmw.TenantValidation(tenantmw.TenantValidationConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		ExcludedRoutes: []string{
			"/health",
			"/api/v1/health",
			"/metrics",
		},
		RejectMissingTenant: true, // cierre de bypass de tenant (rollout verificado 2026-06-19)
	}))

	// Configurar Prometheus metrics si está habilitado
	prometheusEnabled := os.Getenv("PROMETHEUS_ENABLED")
	if prometheusEnabled == "true" {
		log.Println("Registering /metrics endpoint")
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// Configuración de CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "up",
			"service": "payment-method",
		})
	})

	// API v1 group
	apiV1 := router.Group("/api/v1")

	// Setup Payment Method Module
	paymentMethodConfig.SetupPaymentMethodModule(apiV1, db)

	// Iniciar el servidor
	port := env.Get("PORT", "8080")
	log.Printf("Starting Payment Method Service on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func setupDatabase() (*sql.DB, error) {
	host := env.Get("DB_HOST", "localhost")
	port := env.Get("DB_PORT", "5432")
	user := env.Get("DB_USER", "postgres")
	password := env.Get("DB_PASSWORD", "postgres")
	dbname := env.Get("DB_NAME", "payment_method_db")
	sslmode := env.Get("DB_SSLMODE", "disable")

	db, err := postgres.Connect(postgres.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	})
	if err != nil {
		return nil, err
	}

	postgres.StartPoolMonitor(context.Background(), db, postgres.MonitorOptions{
		Service: "payment-method-service",
		DBName:  dbname,
	})

	log.Println("Successfully connected to database")
	return db, nil
}
