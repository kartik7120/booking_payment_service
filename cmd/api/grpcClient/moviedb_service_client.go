package moviedb

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewMovieDBClient() (MovieDBServiceClient, error) {
	// Initialize the MovieDB client here
	// This function should create a new instance of the MovieDB client
	// and return it for use in other parts of the application.
	// You can use environment variables or configuration files to set up the client.
	// Example:
	// client := moviedb.NewClient(moviedb.WithAPIKey(os.Getenv("MOVIEDB_API_KEY")))
	// return client

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(":1102", opts...)

	if err != nil {
		return nil, err
	}

	client := NewMovieDBServiceClient(conn)

	return client, nil
}
