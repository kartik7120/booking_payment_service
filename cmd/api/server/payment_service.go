package server

import (
	moviedb "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
)

type Payment_Service struct {
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

func (m *Payment_Service) Create_Payment_Intent_INR(quantity uint, price uint) (*stripe.PaymentIntent, error) {
	// Create a payment intent with the ticket product
	// Return the payment intent ID

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(price * quantity)),
		Currency: stripe.String(string(stripe.CurrencyINR)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	result, err := paymentintent.New(params)

	if err != nil {
		log.Error("Failed to create payment intent: ", err)
		return nil, err
	}

	log.Info("Payment intent created successfully for amount: ", price*quantity)

	return result, nil
}
