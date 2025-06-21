package server

import (
	"context"
	"os"

	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/go-playground/validator/v10"
	payment_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
)

type Payment_Server struct {
	payment_service.UnimplementedPaymentServiceServer
	Ps *Payment_Service
}

func NewPaymentServer() *Payment_Server {
	// Initialize the Payment_Service
	if os.Getenv("DODOPAYMENT_TOKEN") == "" {
		panic("DODOPAYMENT_TOKEN environment variable is not set")
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
		}
	}

	return &Payment_Server{
		Ps: &Payment_Service{
			Client: dodopayments.NewClient(
				option.WithBearerToken(os.Getenv("DODOPAYMENT_TOKEN")),
			),
			Validator: validator.New(),
		},
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

	paymentLink, err := p.Ps.Create_Payment_Intent_INR(
		CreatePaymentIntentPayload{
			Quantity:    uint(in.Quintity),
			Price:       uint(in.Price),
			Email:       in.Email,
			PhoneNumber: in.PhoneNumber,
			Name:        in.Name,
			MovieName:   in.MovieName,
			Country:     in.Country,
			State:       in.State,
			City:        in.City,
			Street:      in.Street,
			Zipcode:     string(in.Zipcode),
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
