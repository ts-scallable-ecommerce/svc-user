package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/db"
	httptransport "github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/logging"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/rbac"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logging.New()

	dbConn, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer dbConn.Close()

	if cfg.JWTPrivateKeyPath == "" || cfg.JWTPublicKeyPath == "" {
		log.Fatal("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH must be set")
	}

	issuer, err := auth.LoadIssuerFromFiles(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, "svc-user", []string{"users"})
	if err != nil {
		log.Fatalf("failed to load jwt keys: %v", err)
	}

	userRepo := users.NewSQLRepository(dbConn)
	rbacService := rbac.NewService(dbConn)
	userService := users.NewService(userRepo, issuer, rbacService)
	userHandler := handlers.NewUserHandler(userService)

	srv, err := httptransport.NewServer(cfg, logger, issuer, userHandler)
	if err != nil {
		log.Fatalf("failed to create http server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GracefulTimeout)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Printf("error shutting down http server: %v", err)
	}

	log.Println("http server shut down cleanly")
}
