package server

import (
	"context"
	"os"

	payment_service "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	"github.com/stripe/stripe-go/v82"
)

type Payment_Server struct {
	payment_service.UnimplementedPaymentServiceServer
	Ps *Payment_Service
}

func NewPaymentServer() *Payment_Server {
	// Initialize the Payment_Service
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		panic("STRIPE_SECRET_KEY environment variable is not set")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return &Payment_Server{
		Ps: &Payment_Service{},
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

func (p *Payment_Server) CreatePaymentIntentINR(ctx context.Context, in *payment_service.Create_Payment_Intent_INR_Request) (*payment_service.Create_Payment_Intent_INR_Response, error) {

	intent, err := p.Ps.Create_Payment_Intent_INR(uint(in.Quintity), uint(in.Price))

	if err != nil {
		return nil, err
	}

	return &payment_service.Create_Payment_Intent_INR_Response{
		Status:       200,
		Message:      "Payment intent created successfully",
		Error:        "",
		ClientSecret: intent.ClientSecret,
	}, nil
}
