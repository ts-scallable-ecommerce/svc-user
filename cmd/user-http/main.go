package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
	httptransport "github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srv, err := httptransport.NewServer(cfg)
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Printf("error shutting down http server: %v", err)
	}

	log.Println("http server shut down cleanly")
}
