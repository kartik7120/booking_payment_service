package server

import (
	"context"
	"os"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/go-playground/validator/v10"
	moviedb "github.com/kartik7120/booking_payment_service/cmd/api/grpcClient"
	moviedb_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcClient"
	payment_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	log "github.com/sirupsen/logrus"
)

type Payment_Server struct {
	payment_service.UnimplementedPaymentServiceServer
	Ps *Payment_Service
	Ms moviedb_service.MovieDBServiceClient
}

func NewPaymentServer() *Payment_Server {
	// Initialize the Payment_Service
	if os.Getenv("DODOPAYMENT_TOKEN") == "" {
		panic("DODOPAYMENT_TOKEN environment variable is not set")
	}

	moviedb_client, err := moviedb_service.NewMovieDBClient()

	if err != nil {
		log.Errorf("Failed to create MovieDB client: %v", err)
		panic("Failed to create MovieDB client: " + err.Error())
	}

	if os.Getenv("ENV") == "test" {
		return &Payment_Server{
			Ps: &Payment_Service{
				Client: dodopayments.NewClient(
					option.WithBearerToken(os.Getenv("DODOPAYMENT_TOKEN")),
					option.WithBaseURL("https://test.dodopayments.com"),
				),
				Validator: validator.New(),
			},
			Ms: moviedb_client,
		}
	}

	return &Payment_Server{
		Ps: &Payment_Service{
			Client: dodopayments.NewClient(
				option.WithBearerToken(os.Getenv("DODOPAYMENT_TOKEN")),
			),
			Validator: validator.New(),
		},
		Ms: moviedb_client,
	}
}

// Need to implement the CreateCheckoutSession method
func (p *Payment_Server) CreateCheckoutSession(ctx context.Context, in *payment_service.CreateCheckoutSessionRequest) (*payment_service.CreateCheckoutSessionResponse, error) {

	_, err := p.Ps.Create_Checkout_Session(in)

	if err != nil {
		return nil, err
	}

	return &payment_service.CreateCheckoutSessionResponse{
		Status:  200,
		Message: "Checkout session created successfully",
		Error:   "",
	}, nil
}

func (p *Payment_Server) CreatePaymentLink(ctx context.Context, in *payment_service.Create_Payment_Intent_INR_Request) (*payment_service.Create_Payment_Intent_INR_Response, error) {

	// Call the booking_moviedb_service to get the seat price, movie name and other details

	// After details are fetched, create the payment intent

	// Also need to implement the success and failure webhook to update the payment status in the database

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Call the moviedb service to get booked seat details

	// Need to write is valid to commit function to check if seat ids given are valid to purchase / commit as booked for a movie_time_slot and seatMatrixIds

	if err := p.Ps.Validator.Struct(in); err != nil {
		return nil, err
	}

	response, err := p.Ms.IsValidToCommitSeatsForBooking(ctx, &moviedb.IsValidToCommitSeatsForBooking_Request{
		MovieTimeSlotId: in.MovieTimeSlotId,
		SeatMatrixIds:   in.SeatMatrixIDs,
	})

	if response == nil {
		return &payment_service.Create_Payment_Intent_INR_Response{
			Status:  500,
			Error:   err.Error(),
			Message: "",
		}, nil

	}

	if !response.Isvalid {
		return &payment_service.Create_Payment_Intent_INR_Response{
			Status:  400,
			Error:   response.Error,
			Message: "Seats are already booked",
		}, nil
	}

	if err != nil {
		return &payment_service.Create_Payment_Intent_INR_Response{
			Status:  400,
			Error:   response.Error,
			Message: "",
		}, err
	}

	// Call moviedb to get the seat prices as the logic to check if it is valid to commit the seat for booking is already done

	var productBookedSeats []ProductBookedSeats

	for _, v := range response.ToBeBookedSeats {
		productBookedSeats = append(productBookedSeats, ProductBookedSeats{
			BookedSeatsID: uint(v.Id),
			Quantity:      1,
			Price:         uint(v.Price),
			SeatNumber:    v.SeatNumber,
			MovieName:     v.MovieName,
		})
	}

	paymentLink, err := p.Ps.Create_Payment_Intent_INR(
		CreatePaymentIntentPayload{
			Email:       in.Email,
			PhoneNumber: in.PhoneNumber,
			Country:     in.Country,
			State:       in.State,
			City:        in.City,
			Street:      in.Street,
			Zipcode:     string(in.Zipcode),
			Name:        in.CustomerName,
			Products:    productBookedSeats,
		},
	)

	if err != nil {
		return nil, err
	}

	return &payment_service.Create_Payment_Intent_INR_Response{
		Status:      200,
		Message:     "Payment intent created successfully",
		Error:       "",
		PaymentLink: paymentLink,
	}, nil
}
