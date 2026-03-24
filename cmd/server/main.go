package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	_ "github.com/sssseraphim/effective_mobile/docs"
	"github.com/sssseraphim/effective_mobile/internal/config"
	"github.com/sssseraphim/effective_mobile/internal/handlers"
	"github.com/sssseraphim/effective_mobile/internal/repository"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	repo, err := repository.NewRepository(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer repo.Close()

	// Run migrations
	migrateURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	log.WithField("migrate_url", migrateURL).Debug("Migration URL")

	// Запускаем миграции
	m, err := migrate.New("file://migrations", migrateURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to create migrate instance")
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.WithError(err).Fatal("Failed to run migrations")
	}
	log.Info("Database migrations completed")

	handler := handlers.NewSubscriptionHandler(repo)

	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	{
		api.POST("/subscriptions", handler.CreateSubscription)
		api.GET("/subscriptions", handler.ListSubscriptions)
		api.GET("/subscriptions/total-cost", handler.GetTotalCost)
		api.GET("/subscriptions/:id", handler.GetSubscription)
		api.PUT("/subscriptions/:id", handler.UpdateSubscription)
		api.DELETE("/subscriptions/:id", handler.DeleteSubscription)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Infof("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
