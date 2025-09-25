package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9090"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	log.Printf("gRPC server listening on %s", addr)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
