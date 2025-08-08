package server

import (
	"context"
	"os"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/go-playground/validator/v10"
	moviedb_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcClient"
	payment_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	if os.Getenv("DB_URL_TEST") == "" {
		panic("DB_URL environment variable is not set")
	}

	moviedb_client, err := moviedb_service.NewMovieDBClient()

	if err != nil {
		log.Errorf("Failed to create MovieDB client: %v", err)
		panic("Failed to create MovieDB client: " + err.Error())
	}

	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_URL_TEST")), &gorm.Config{})

	if err != nil {
		log.Errorf("Failed to connect to the database: %v", err)
		panic("Failed to connect to the database: " + err.Error())
	}

	if os.Getenv("ENV") == "test" {
		return &Payment_Server{
			Ps: &Payment_Service{
				Client: dodopayments.NewClient(
					option.WithBearerToken(os.Getenv("DODOPAYMENT_TOKEN")),
					option.WithBaseURL("https://test.dodopayments.com"),
				),
				Validator: validator.New(),
				DB:        conn,
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
			DB:        conn,
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

	response, err := p.Ms.IsValidToCommitSeatsForBooking(ctx, &moviedb_service.IsValidToCommitSeatsForBooking_Request{
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

func (p *Payment_Server) IsValidIdempotentKey(ctx context.Context, in *payment_service.IsValidIdempotentKeyRequest) (*payment_service.IsValidIdempotentKeyResponse, error) {
	// Check if the idempotent key is valid
	isValid, err := p.Ps.IsValidateItempotentKey(in.IdempotentKey)

	if err != nil {
		return &payment_service.IsValidIdempotentKeyResponse{
			IsValid: false,
			Error:   err.Error(),
			Message: "Failed to validate idempotent key",
		}, nil
	}

	return &payment_service.IsValidIdempotentKeyResponse{
		IsValid: isValid,
		Error:   "",
		Message: "Idempotent key is valid",
	}, nil
}

func (p *Payment_Server) CommitIdempotentKey(ctx context.Context, in *payment_service.CommitIdempotentKeyRequest) (*payment_service.Create_Payment_Intent_INR_Response, error) {
	// Commit the idempotent key
	err := p.Ps.CommitIdempotentKey(in.IdempotentKey)

	if err != nil {
		return &payment_service.Create_Payment_Intent_INR_Response{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to commit idempotent key",
		}, nil
	}

	return &payment_service.Create_Payment_Intent_INR_Response{
		Status:  200,
		Error:   "",
		Message: "Idempotent key committed successfully",
	}, nil
}

func (p *Payment_Server) CreateOrder(ctx context.Context, in *payment_service.Create_Order_Request) (*payment_service.Create_Order_Response, error) {

	var productIds []string

	if in.MovieTimeSlotId == 0 || len(in.SeatMatrixIDs) == 0 {
		return &payment_service.Create_Order_Response{
			Status:  400,
			Error:   "Movie time slot ID and seat matrix IDs cannot be empty",
			Message: "Failed to create order",
		}, nil
	}

	// Validate the idempotent key

	if in.IdempotentKey == "" {
		return &payment_service.Create_Order_Response{
			Status:  400,
			Error:   "Idempotent key cannot be empty",
			Message: "Failed to create order",
		}, nil
	}

	// Call the moviedb service to check if the movie time slot ID and seat matrix IDs are valid

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	response, err := p.Ms.IsValidToCommitSeatsForBooking(ctx, &moviedb_service.IsValidToCommitSeatsForBooking_Request{
		MovieTimeSlotId: in.MovieTimeSlotId,
		SeatMatrixIds:   in.SeatMatrixIDs,
	})

	if err != nil {
		return &payment_service.Create_Order_Response{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to validate movie time slot and seat matrix IDs",
		}, nil
	}

	if response == nil || !response.Isvalid {
		return &payment_service.Create_Order_Response{
			Status:  400,
			Error:   response.Error,
			Message: "Invalid movie time slot or seat matrix IDs",
		}, nil
	}

	// Call the moviedb service to get information about the movie name, seats that need to be booked, and their prices

	for _, order := range response.ToBeBookedSeats {
		// Create a product for each booked seat

		product, err := p.Ps.Create_Product_Ticket(Product{
			ProductName:        order.MovieName + " - " + order.SeatNumber,
			Price:              int64(order.Price),
			ProductDescription: "Seat " + order.SeatNumber + " for movie " + order.MovieName,
		})

		if err != nil {
			return &payment_service.Create_Order_Response{
				Status:  500,
				Error:   err.Error(),
				Message: "Failed to create product",
			}, nil
		}

		productIds = append(productIds, product.ProductID)

	}

	// Once orders are created, commit the order IDs to the idempotent key

	var bookedSeatsID []int32

	for _, v := range response.ToBeBookedSeats {
		bookedSeatsID = append(bookedSeatsID, v.Id)
	}

	err = p.Ps.CommitOrderIDs(in.IdempotentKey, productIds, int(in.MovieTimeSlotId), bookedSeatsID)

	if err != nil {
		return &payment_service.Create_Order_Response{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to commit order IDs",
		}, nil
	}

	return &payment_service.Create_Order_Response{
		Status:  200,
		Error:   "",
		Message: "Order created successfully",
		OrderId: productIds,
	}, nil
}

func (p *Payment_Server) CommitCustomerID(ctx context.Context, in *payment_service.CommitIdempotentKeyRequest) (*payment_service.Create_Payment_Intent_INR_Response, error) {

	// Commit the customer ID to the idempotent key
	err := p.Ps.CommitCustomerPaymentSession(in.IdempotentKey, in.CustomerId)

	if err != nil {
		return &payment_service.Create_Payment_Intent_INR_Response{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to commit customer ID",
		}, nil
	}

	return &payment_service.Create_Payment_Intent_INR_Response{
		Status:  200,
		Error:   "",
		Message: "Customer ID committed successfully",
	}, nil
}

func (p *Payment_Server) CommitOrderIds(ctx context.Context, in *payment_service.CommitIdempotentKeyRequest) (*payment_service.Create_Payment_Intent_INR_Response, error) {

	// Commit the order IDs to the idempotent key
	err := p.Ps.CommitOrderIDs(in.IdempotentKey, in.OrderIds, 0, []int32{})

	if err != nil {
		return &payment_service.Create_Payment_Intent_INR_Response{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to commit order IDs",
		}, nil
	}

	return &payment_service.Create_Payment_Intent_INR_Response{
		Status:  200,
		Error:   "",
		Message: "Order IDs committed successfully",
	}, nil
}

func (p *Payment_Server) CreateCustomer(ctx context.Context, in *payment_service.CreateCustomerRequest) (*payment_service.CreateCustomerResponse, error) {
	// Create a customer in the database with the given details
	customer, err := p.Ps.Create_Customer(in.Email, in.CustomerName, in.PhoneNumber, in.IdempotentKey)

	if err != nil {
		return &payment_service.CreateCustomerResponse{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to create customer",
		}, nil
	}

	return &payment_service.CreateCustomerResponse{
		Status:     200,
		Error:      "",
		Message:    "Customer created successfully",
		CustomerId: customer.CustomerID,
	}, nil
}

func (p *Payment_Server) GeneratePaymentLink(ctx context.Context, in *payment_service.CreatePaymentLinkRequest) (*payment_service.CreatePaymentLinkResponse, error) {
	// Generate a payment link for the customer
	paymentLink, err := p.Ps.GeneratePaymentLink(in.IdempotentKey)

	if err != nil {
		return &payment_service.CreatePaymentLinkResponse{
			Status:  500,
			Error:   err.Error(),
			Message: "Failed to generate payment link",
		}, nil
	}

	return &payment_service.CreatePaymentLinkResponse{
		Status:      200,
		Error:       "",
		Message:     "Payment link generated successfully",
		PaymentLink: paymentLink,
	}, nil
}
