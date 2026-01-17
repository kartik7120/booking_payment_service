package moviedb

import (
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func NewMovieDBClient() (MovieDBServiceClient, error) {
	// Initialize the MovieDB client here
	// This function should create a new instance of the MovieDB client
	// and return it for use in other parts of the application.
	// You can use environment variables or configuration files to set up the client.
	// Example:
	// client := moviedb.NewClient(moviedb.WithAPIKey(os.Getenv("MOVIEDB_API_KEY")))
	// return client

	// var opts []grpc.DialOption
	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	moviedb_service_url := os.Getenv("MOVIEDB_URL")

	if moviedb_service_url == "" {
		moviedb_service_url = "booking_moviedb_service:1102"
	}

	conn, err := grpc.NewClient(moviedb_service_url, grpcOpts...)

	if err != nil {
		return nil, err
	}

	client := NewMovieDBServiceClient(conn)

	return client, nil
}
