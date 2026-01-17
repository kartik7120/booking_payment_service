package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	payment_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	"github.com/kartik7120/booking_payment_service/cmd/api/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {

	// --------------------------------------------------
	// ENV (load .env only locally)
	// --------------------------------------------------
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}

	// --------------------------------------------------
	// Logging
	// --------------------------------------------------
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)

	if os.Getenv("ENV") == "production" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	// --------------------------------------------------
	// Port (Cloud Run SAFE)
	// --------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "1104" // local fallback ONLY
	}

	addr := "0.0.0.0:" + port
	log.Infof("Starting Payment gRPC service on %s", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	// --------------------------------------------------
	// gRPC Server
	// --------------------------------------------------
	grpcServer := grpc.NewServer()

	paymentServer := server.NewPaymentServer()
	payment_service.RegisterPaymentServiceServer(grpcServer, paymentServer)

	// --------------------------------------------------
	// Graceful shutdown
	// --------------------------------------------------
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info("Payment gRPC server is now listening")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	<-signalChan
	log.Info("Shutdown signal received")

	grpcServer.GracefulStop()
	log.Info("Payment service stopped gracefully")
}
