package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	payment_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	"github.com/kartik7120/booking_payment_service/cmd/api/server"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Error("Error loading .env file")
		panic(err)
	}
	log.SetOutput(os.Stdout)
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	lis, err := net.Listen("tcp", ":1104")

	if err != nil {
		log.Error("Error starting the server")
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	// Register the service here

	paymentServer := server.NewPaymentServer()

	payment_service.RegisterPaymentServiceServer(grpcServer, paymentServer)

	log.Info("Payment Service is running on port 1104")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("Failed to start gRPC server: ", err)
			os.Exit(1)
		}
	}()

	<-signalChan

	log.Info("Received shutdown signal, stopping server...")

	grpcServer.GracefulStop()

	log.Info("Server stopped gracefully")

	if err := lis.Close(); err != nil {
		log.Error("Failed to close listener: ", err)
	} else {
		log.Info("Listener closed successfully")
	}

	log.Info("Payment Service has stopped")

}
