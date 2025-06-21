package server

import (
	"context"
	"fmt"

	"github.com/dodopayments/dodopayments-go"
	"github.com/go-playground/validator/v10"
	moviedb "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	log "github.com/sirupsen/logrus"
)

type Payment_Service struct {
	Client    *dodopayments.Client
	Validator *validator.Validate
}

type CreatePaymentIntentPayload struct {
	Quantity    uint   `json:"quantity" validate:"required"`
	Price       uint   `json:"price" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Name        string `json:"name" validate:"required"`
	MovieName   string `json:"movie_name" validate:"required"`
	Country     string `json:"country" validate:"required"`
	State       string `json:"state" validate:"required"`
	City        string `json:"city" validate:"required"`
	Street      string `json:"street" validate:"required"`
	Zipcode     string `json:"zipcode" validate:"required"`
}

func (m *Payment_Service) Create_Checkout_Session(request *moviedb.CreateCheckoutSessionRequest) (int, error) {
	// Create a ticket product
	// Create a checkout session with the ticket product
	// Return the session ID

	// productParams := &stripe.ProductParams{
	// 	Name: stripe.String(),
	// 	Type: stripe.String(string(stripe.ProductTypeService)),
	// 	Metadata: map[string]string{
	// 		"movie_name": movieName,
	// 	},
	// }

	// _, err := product.New(productParams)

	// if err != nil {
	// 	log.Error("Failed to create product: ", err)
	// 	return err
	// }

	// log.Info("Product created successfully for movie: ", movieName)

	// // Customize the product using price API

	// params := &stripe.CheckoutSessionParams{
	// 	Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	// 	LineItems: []*stripe.CheckoutSessionLineItemParams{
	// 		&stripe.CheckoutSessionLineItemParams{},
	// 	},
	// }

	// return nil

	return 200, nil
}

func (m *Payment_Service) Create_Payment_Intent_INR(payload CreatePaymentIntentPayload) (string, error) {

	customer, err := m.Create_Customer(payload.Email, payload.Name, payload.PhoneNumber)

	if err != nil {
		return "", err
	}

	Product, err := m.Create_Product_Ticket(Product{
		ProductName:        payload.MovieName,
		Price:              int64(payload.Price),
		ProductDescription: fmt.Sprintf("Ticket for movie: %s", payload.MovieName),
	})

	if err != nil {
		return "", err
	}

	log.Infof("Creating payment intent for product: %s with price: %d and quintity: %d", Product.ProductID, payload.Price, payload.Quantity)

	dodopayments.NewWebhookEventService()
	payment, err := m.Client.Payments.New(
		context.TODO(),
		dodopayments.PaymentNewParams{
			PaymentLink: dodopayments.F(true),
			ProductCart: dodopayments.F(
				[]dodopayments.OneTimeProductCartItemParam{
					{
						Quantity:  dodopayments.F(int64(payload.Quantity)),
						ProductID: dodopayments.F(Product.ProductID), // Replace with actual product ID
					},
				},
			),
			Billing: dodopayments.F(
				dodopayments.BillingAddressParam{
					Country: dodopayments.F(dodopayments.CountryCodeIn),
					State:   dodopayments.F(payload.State),
					City:    dodopayments.F(payload.City),
					Street:  dodopayments.F(payload.Street),
					Zipcode: dodopayments.F(payload.Zipcode),
				},
			),
			Customer: dodopayments.F[dodopayments.CustomerRequestUnionParam](dodopayments.AttachExistingCustomerParam{
				CustomerID: dodopayments.F(customer.CustomerID),
			}),
			ReturnURL:       dodopayments.F("https://example.com/return"), // Replace with your return URL
			BillingCurrency: dodopayments.F(dodopayments.CurrencyInr),
		},
	)

	if err != nil {
		log.Error("Failed to create payment intent: ", err)
		return "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	return payment.PaymentLink, nil
}
