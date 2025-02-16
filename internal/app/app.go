package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"avito-internship/config"
	"avito-internship/internal/cache/redis"
	v1 "avito-internship/internal/controller/http/v1"
	"avito-internship/internal/repository"
	"avito-internship/internal/service"
	"avito-internship/pkg/logger"
	"avito-internship/pkg/postgres"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func Run() {
	// Config init
	cfg := config.MustLoad()

	// Logger init
	log := logger.NewZap(cfg.Env)

	// Context
	ctx, cancel := context.WithCancel(context.Background())

	// Database init
	log.Info("Database initialization...")
	pg := postgres.NewPostgres(ctx, log, cfg.PgDSN)
	log.Info("Database initialization: OK.")

	// Cache init
	log.Info("Cache initialization...")
	cache := redis.NewRedisCache(cfg.RedisDSN, log)
	log.Info("Cache initialization: OK.")

	// Repositories init
	log.Info("Repository initialization...")
	repositories := repository.NewRepositories(pg)
	log.Info("Repository initialization: OK.")

	// Services init
	log.Info("Services initialization...")
	deps := service.ServicesDependencies{
		Log:      log,
		Cache:    cache,
		Repos:    repositories,
		TokenTTL: cfg.TokenTTL,
		Salt:     cfg.Salt,
	}
	services := service.NewServices(deps)
	log.Info("Services initialization: OK.")

	// Channel for signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Router init
	log.Info("Router initialization...")
	app := fiber.New(fiber.Config{
		AppName: "API Avito shop",
	})
	app.Use(recover.New(recover.Config{
		EnableStackTrace: false,
	}))
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	middleware := v1.NewAuthMiddleware(log, services.Auth)
	v1.InitRouter(ctx, log, app, services, middleware.Auth())
	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Error("Fiber server error",
				zap.Error(err),
			)
		}
	}()

	// Graceful shutdown
	<-quit
	log.Info("Shutdown signal received")
	cancel()

	if err := app.Shutdown(); err != nil {
		log.Error("Error shutting down Fiber",
			zap.Error(err),
		)
	} else {
		log.Info("Fiber server stopped")
	}

	if pg.Pool != nil {
		pg.Close()
	}
	log.Info("Database connection closed")

	if err := cache.Shutdown(); err != nil {
		log.Error("Failed to close Redis client",
			zap.Error(err),
		)
	}
	log.Info("Redis connection closed")

	log.Info("Gracefully stopped")
}
