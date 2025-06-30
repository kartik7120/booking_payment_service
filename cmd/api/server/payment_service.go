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

type ProductBookedSeats struct {
	BookedSeatsID uint   `json:"booked_seats_id" gorm:"not null"`
	Quantity      uint   `json:"quantity" gorm:"not null;default:1"` // number of tickets booked
	Price         uint   `json:"price" gorm:"not null"`              // price per ticket
	SeatNumber    string `json:"seat_number" gorm:"not null"`        // seat number booked
	MovieName     string `json:"movie_name" gorm:"not null"`         // name of the movie for which the seat is booked
}

type CreatePaymentIntentPayload struct {
	Products []ProductBookedSeats `json:"product" validate:"required,dive,required"`
	// Price       uint   `json:"price" validate:"required"` price will be given by fetching it from the booking_moviedb_service
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Name        string `json:"name" validate:"required"`
	// MovieName   string `json:"movie_name" validate:"required"` movie name will be given by fetching it from the booking_moviedb_service
	Country string `json:"country" validate:"required"`
	State   string `json:"state" validate:"required"`
	City    string `json:"city" validate:"required"`
	Street  string `json:"street" validate:"required"`
	Zipcode string `json:"zipcode" validate:"required"`
}

// Need to create a webhook to handle payment success and failure events
// and update the payment status in the database accordingly

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

	log.Infof("Email %s, Name: %s and phone_number: %s", payload.Email, payload.Name, payload.PhoneNumber)

	customer, err := m.Create_Customer(payload.Email, payload.Name, payload.PhoneNumber)

	if err != nil {
		return "", err
	}

	var ProductArr []*dodopayments.Product

	// Product, err := m.Create_Product_Ticket(Product{
	// 	ProductName:        payload.MovieName,
	// 	Price:              int64(payload.Price),
	// 	ProductDescription: fmt.Sprintf("Ticket for movie: %s", payload.MovieName),
	// })

	// First check if these seats are already booked or exist in the database

	for _, v := range payload.Products {

		Product, err := m.Create_Product_Ticket(Product{
			ProductName:        v.MovieName,
			Price:              int64(v.Price),
			ProductDescription: fmt.Sprintf("Ticket for movie: %s", v.MovieName),
		})

		log.Infof("Creating payment intent for product: %s with price: %d and quintity: %d", Product.ProductID, v.Price, v.Quantity)

		if err != nil {
			log.Error("Failed to create product: ", err)
			return "", fmt.Errorf("failed to create product: %w", err)
		}

		ProductArr = append(ProductArr, Product)
	}

	ProductCartItems := make([]dodopayments.OneTimeProductCartItemParam, len(ProductArr))

	for i, product := range ProductArr {
		ProductCartItems[i] = dodopayments.OneTimeProductCartItemParam{
			ProductID: dodopayments.F(product.ProductID),
			Quantity:  dodopayments.F(int64(1)), // Assuming quantity is always 1 for each product
		}
	}

	dodopayments.NewWebhookEventService()
	payment, err := m.Client.Payments.New(
		context.TODO(),
		dodopayments.PaymentNewParams{
			PaymentLink: dodopayments.F(true),
			ProductCart: dodopayments.F(ProductCartItems),
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

	// confimation of the booked seats will be handled by the webhooks

	if err != nil {
		log.Error("Failed to create payment intent: ", err)
		return "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	return payment.PaymentLink, nil
}
