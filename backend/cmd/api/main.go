// Setor.in API server entry point.
//
// This is the composition root: it loads config, wires shared services
// (logger, DB pool), starts the HTTP server, and handles graceful shutdown.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	authpersist "github.com/setorin/setorin/backend/internal/apps/auth/adapters/persistence"
	authhttp "github.com/setorin/setorin/backend/internal/apps/auth/delivery/http"
	authuc "github.com/setorin/setorin/backend/internal/apps/auth/application/usecase"
	catalogpersist "github.com/setorin/setorin/backend/internal/apps/catalog/adapters/persistence"
	cataloghttp "github.com/setorin/setorin/backend/internal/apps/catalog/delivery/http"
	cataloguc "github.com/setorin/setorin/backend/internal/apps/catalog/application/usecase"
	orderpersist "github.com/setorin/setorin/backend/internal/apps/order/adapters/persistence"
	orderhttp "github.com/setorin/setorin/backend/internal/apps/order/delivery/http"
	orderuc "github.com/setorin/setorin/backend/internal/apps/order/application/usecase"
	userpersist "github.com/setorin/setorin/backend/internal/apps/user/adapters/persistence"
	userhttp "github.com/setorin/setorin/backend/internal/apps/user/delivery/http"
	useruc "github.com/setorin/setorin/backend/internal/apps/user/application/usecase"
	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	"github.com/setorin/setorin/backend/internal/services/email"
	"github.com/setorin/setorin/backend/internal/shared/config"
	"github.com/setorin/setorin/backend/internal/shared/database"
	"github.com/setorin/setorin/backend/internal/shared/logger"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

func main() {
	if err := run(); err != nil {
		// run() already logs; this is the last-chance fallback before exit.
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.New(cfg.App.LogLevel)
	slog.Info("starting setorin-api",
		"env", cfg.App.Env,
		"port", cfg.App.Port,
		"version", "0.1.0",
	)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.New(rootCtx, cfg.DB)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer pool.Close()

	jwtValidator, err := keycloak.NewValidator(rootCtx, *cfg)
	if err != nil {
		return fmt.Errorf("init JWT validator: %w", err)
	}

	kcAdmin := keycloak.NewAdminClient(cfg.Keycloak)
	if err := kcAdmin.HealthCheck(rootCtx); err != nil {
		// Non-fatal at startup — log and continue. Admin operations will retry.
		slog.Warn("keycloak admin health check failed; will retry on demand", "err", err)
	}

	emailSvc, err := email.NewService(cfg.SMTP)
	if err != nil {
		return fmt.Errorf("init email service: %w", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())
	r.Use(corsMiddleware(cfg.CORS.AllowedOrigins))

	registerRoutes(r, pool, jwtValidator, kcAdmin, emailSvc)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("http server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "err", err)
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-stop:
		slog.Info("shutdown signal received", "signal", sig.String())
	case <-rootCtx.Done():
		slog.Warn("root context cancelled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		return err
	}

	slog.Info("server stopped cleanly")
	return nil
}

// registerRoutes wires up all HTTP routes. Each module owns its own
// RegisterRoutes function that mounts under the shared /v1 group.
func registerRoutes(
	r *gin.Engine,
	pool *pgxpool.Pool,
	jwtV *keycloak.Validator,
	kcAdmin *keycloak.AdminClient,
	mail *email.Service,
) {
	// ───── Public ─────
	r.GET("/health", healthHandler(pool))
	r.GET("/", rootHandler())

	// ───── /v1 group ─────
	v1 := r.Group("/v1")

	// Shared repositories
	userRepo := userpersist.NewPostgresUserRepo(pool)
	auditRepo := authpersist.NewPostgresAuthEventRepo(pool)
	collectorRepo := authpersist.NewPostgresCollectorAppRepo(pool)
	adminUserRepo := authpersist.NewPostgresAdminUserRepo(pool)

	// Shared services
	auditSvc := authuc.NewAuditService(auditRepo)

	// User module — /v1/auth/sync, /v1/auth/me, /v1/users/me*
	userUC := useruc.NewUserUseCase(userRepo, kcAdmin, mail)
	userHandler := userhttp.NewHandler(userUC)
	userhttp.RegisterRoutes(v1, userHandler, jwtV)

	// Auth module — /v1/users/me/become-collector, /v1/admin/*
	collectorUC := authuc.NewCollectorApplicationUseCase(collectorRepo, userRepo, adminUserRepo, kcAdmin, auditSvc, mail)
	adminUC := authuc.NewAdminUserUseCase(adminUserRepo, kcAdmin, auditSvc, mail)
	adminHandler := authhttp.NewAdminHandler(adminUC, collectorUC, userRepo)
	userCollectorHandler := authhttp.NewUserCollectorHandler(collectorUC)
	authhttp.RegisterRoutes(v1, adminHandler, userCollectorHandler, jwtV)

	// Catalog module — public /v1/materials, admin /v1/admin/materials
	materialRepo := catalogpersist.NewPostgresMaterialRepo(pool)
	materialUC := cataloguc.NewMaterialUseCase(materialRepo)
	materialHandler := cataloghttp.NewHandler(materialUC, userRepo)
	cataloghttp.RegisterRoutes(v1, materialHandler, jwtV)

	// Order module — /v1/orders/*, /v1/collector/orders/*
	orderRepo := orderpersist.NewPostgresOrderRepo(pool)
	userOrderUC := orderuc.NewUserOrderUseCase(orderRepo, materialRepo, userRepo, mail)
	collectorOrderUC := orderuc.NewCollectorOrderUseCase(orderRepo, userRepo, mail)
	userOrderHandler := orderhttp.NewUserHandler(userOrderUC, userRepo)
	collectorOrderHandler := orderhttp.NewCollectorHandler(collectorOrderUC, userRepo)
	orderhttp.RegisterRoutes(v1, userOrderHandler, collectorOrderHandler, jwtV)

	// ───── Smoke-test endpoint kept for RBAC verification ─────
	v1.GET("/admin/ping",
		middleware.JWT(jwtV),
		middleware.RequireRole(middleware.RoleAdmin),
		func(c *gin.Context) {
			claims := middleware.MustClaims(c)
			response.OK(c, gin.H{
				"message": "hello, admin",
				"sub":     claims.Subject,
			})
		},
	)
}

func healthHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "down"
		}

		response.OK(c, gin.H{
			"status": "ok",
			"checks": gin.H{"db": dbStatus},
		})
	}
}

func rootHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.OK(c, gin.H{
			"name":    "setorin-api",
			"version": "0.1.0",
		})
	}
}

// requestLogger logs each request with method, path, status, and duration.
// It redacts the Authorization header (defense in depth).
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		slog.Info("http",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	if len(origins) == 0 {
		// CORS misconfigured: deny all browser cross-origin requests.
		// (server-to-server clients are unaffected since they don't send Origin.)
		return func(c *gin.Context) { c.Next() }
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
