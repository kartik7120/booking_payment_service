package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/go-playground/validator/v10"
	moviedb "github.com/kartik7120/booking_payment_service/cmd/api/grpcServer"
	"github.com/kartik7120/booking_payment_service/cmd/api/models"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Payment_Service struct {
	Client    *dodopayments.Client
	Validator *validator.Validate
	DB        *gorm.DB
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
	Country       string `json:"country" validate:"required"`
	State         string `json:"state" validate:"required"`
	City          string `json:"city" validate:"required"`
	Street        string `json:"street" validate:"required"`
	Zipcode       string `json:"zipcode" validate:"required"`
	IdempotentKey string `json:"idempotent_key" validate:"required"`
}

type PaymentDetail struct {
	Billing struct {
		City    string `json:"city"`
		Country string `json:"country"`
		State   string `json:"state"`
		Street  string `json:"street"`
		Zipcode string `json:"zipcode"`
	} `json:"billing"`

	BrandID            string  `json:"brand_id"`
	BusinessID         string  `json:"business_id"`
	CardIssuingCountry *string `json:"card_issuing_country"`
	CardLastFour       string  `json:"card_last_four"`
	CardNetwork        string  `json:"card_network"`
	CardType           string  `json:"card_type"`
	CreatedAt          string  `json:"created_at"`
	Currency           string  `json:"currency"`
	Customer           struct {
		CustomerID string `json:"customer_id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
	} `json:"customer"`
	DigitalProductsDelivered bool   `json:"digital_products_delivered"`
	DiscountID               string `json:"discount_id"`

	Disputes []struct {
		Amount        string `json:"amount"`
		BusinessID    string `json:"business_id"`
		CreatedAt     string `json:"created_at"`
		Currency      string `json:"currency"`
		DisputeID     string `json:"dispute_id"`
		DisputeStage  string `json:"dispute_stage"`
		DisputeStatus string `json:"dispute_status"`
		PaymentID     string `json:"payment_id"`
		Remarks       string `json:"remarks"`
	} `json:"disputes"`

	ErrorCode         string                 `json:"error_code"`
	ErrorMessage      string                 `json:"error_message"`
	Metadata          map[string]interface{} `json:"metadata"`
	PaymentID         string                 `json:"payment_id"`
	PaymentLink       string                 `json:"payment_link"`
	PaymentMethod     string                 `json:"payment_method"`
	PaymentMethodType string                 `json:"payment_method_type"`
	ProductCart       []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"product_cart"`

	Refunds []struct {
		Amount     int     `json:"amount"`
		BusinessID string  `json:"business_id"`
		CreatedAt  string  `json:"created_at"`
		Currency   *string `json:"currency"`
		IsPartial  bool    `json:"is_partial"`
		PaymentID  string  `json:"payment_id"`
		Reason     string  `json:"reason"`
		RefundID   string  `json:"refund_id"`
		Status     string  `json:"status"`
	} `json:"refunds"`

	SettlementAmount   int     `json:"settlement_amount"`
	SettlementCurrency string  `json:"settlement_currency"`
	SettlementTax      int     `json:"settlement_tax"`
	Status             *string `json:"status"`
	SubscriptionID     string  `json:"subscription_id"`
	Tax                int     `json:"tax"`
	TotalAmount        int     `json:"total_amount"`
	UpdatedAt          string  `json:"updated_at"`
}

type CustomerDetail struct {
	BusinessID  string `json:"business_id"`
	CreatedAt   string `json:"created_at"` // Consider time.Time if using custom unmarshalling
	CustomerID  string `json:"customer_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type ProductDetail struct {
	Addons                     []string `json:"addons"`
	BrandID                    string   `json:"brand_id"`
	BusinessID                 string   `json:"business_id"`
	CreatedAt                  string   `json:"created_at"` // use time.Time with custom parsing if needed
	Description                string   `json:"description"`
	DigitalProductDelivery     *string  `json:"digital_product_delivery"` // nullable
	Image                      string   `json:"image"`
	IsRecurring                bool     `json:"is_recurring"`
	LicenseKeyActivationMsg    string   `json:"license_key_activation_message"`
	LicenseKeyActivationsLimit int      `json:"license_key_activations_limit"`
	LicenseKeyDuration         *string  `json:"license_key_duration"` // nullable
	LicenseKeyEnabled          bool     `json:"license_key_enabled"`
	Name                       string   `json:"name"`
	Price                      struct {
		Currency              string `json:"currency"`
		Discount              int    `json:"discount"`
		PayWhatYouWant        bool   `json:"pay_what_you_want"`
		Price                 int    `json:"price"`
		PurchasingPowerParity bool   `json:"purchasing_power_parity"`
		SuggestedPrice        int    `json:"suggested_price"`
		TaxInclusive          bool   `json:"tax_inclusive"`
		Type                  string `json:"type"`
	} `json:"price"`
	ProductID   string `json:"product_id"`
	TaxCategory string `json:"tax_category"`
	UpdatedAt   string `json:"updated_at"` // could be time.Time
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

	// Validate the input payload

	if err := m.Validator.Struct(payload); err != nil {
		log.Error("Validation failed for CreatePaymentIntentPayload: ", err)
		return "", fmt.Errorf("validation failed: %w", err)
	}

	customer, err := m.Create_Customer(payload.Email, payload.Name, payload.PhoneNumber, payload.IdempotentKey)

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

func (m *Payment_Service) IsValidateItempotentKey(key string) (bool, error) {

	var idempotent models.Idempotent

	result := m.DB.Where("idempotent_key = ?", key).First(&idempotent)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Infof("Idempotent key %s not found", key)
			return false, nil // Key not found, so it's valid to use
		}
		log.Error("Error checking idempotent key: ", result.Error)
		return false, fmt.Errorf("error checking idempotent key: %w", result.Error)
	}

	log.Infof("Idempotent key %s found, created at %s", key, fmt.Sprint(idempotent.CreatedAt))

	return true, nil // Key found, so it's not valid to use again
}

func (m *Payment_Service) CommitIdempotentKey(key string) error {

	result := m.DB.Model(&models.Idempotent{}).Create(&models.Idempotent{
		IdempotentKey: key,
	})

	if result.Error != nil {
		log.Error("Error committing idempotent key: ", result.Error)

		if result.Error == gorm.ErrDuplicatedKey {
			log.Infof("Idempotent key %s already exists, skipping commit", key)
			return nil // Key already exists, so we can skip committing it again
		}

		return fmt.Errorf("error committing idempotent key: %w", result.Error)
	}

	log.Infof("Idempotent key %s committed successfully for customer", key)

	return nil
}

func (m *Payment_Service) CommitCustomerPaymentSession(key string, customerID string) error {

	// Add customer id to the idempotent table

	result := m.DB.Model(&models.Idempotent{}).Where("idempotent_key = ?", key).Updates(models.Idempotent{
		CustomerID: customerID,
	})

	if result.Error != nil {
		log.Error("Error committing customer payment session: ", result.Error)
		return fmt.Errorf("error committing customer payment session: %w", result.Error)
	}

	log.Infof("Customer payment session committed successfully for idempotent key %s with customer ID %s", key, customerID)

	return nil
}

func (c *Payment_Service) CommitOrderIDs(key string, orderIDs []string, movieTimeSlotID int, bookedSeatsID []int32) error {

	// Add order IDs to the idempotent table
	log.Infof("Committing order IDs for idempotent key %s", key)
	log.Infof("Order IDs: %v, Movie Time Slot ID: %d, Booked Seats ID: %v", orderIDs, movieTimeSlotID, bookedSeatsID)

	var orderIds pq.StringArray

	for _, orderID := range orderIDs {
		orderIds = append(orderIds, orderID)
	}

	var bookedSeatsId pq.Int32Array

	for _, seatID := range bookedSeatsID {
		bookedSeatsId = append(bookedSeatsId, int32(seatID))
	}

	result := c.DB.Model(&models.Idempotent{}).Where("idempotent_key = ?", key).Updates(&models.Idempotent{
		OrderIDs:        orderIds,
		MovieTimeSlotID: uint(movieTimeSlotID),
		BookedSeatsId:   bookedSeatsId,
		PaymentStatus:   "INITIATED",
	})

	if result.Error != nil {
		log.Error("Error committing order IDs: ", result.Error)
		return fmt.Errorf("error committing order IDs: %w", result.Error)
	}

	log.Infof("Order IDs committed successfully for idempotent key %s", key)

	return nil
}

func (c *Payment_Service) GeneratePaymentLink(idempotentKey string) (string, error) {

	log.Infof("Generating payment link for idempotent key: %s", idempotentKey)

	// Check if the idempotent key exists
	exists, err := c.IsValidateItempotentKey(idempotentKey)
	if err != nil {
		return "", fmt.Errorf("error validating idempotent key: %w", err)
	}

	// If the key does not exist then just return an error

	if !exists {
		return "", fmt.Errorf("idempotent key %s does not exist", idempotentKey)
	}

	// Commit the idempotent key to the database

	// Use the customer id to get customer details

	// Use movie time slot ID to get movie details

	// Use booked seats ID to get seat details

	var Idempotent models.Idempotent

	result := c.DB.Where("idempotent_key = ?", idempotentKey).First(&Idempotent)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Errorf("Idempotent key %s not found", idempotentKey)
			return "", fmt.Errorf("idempotent key %s not found", idempotentKey)
		}
		log.Error("Error fetching idempotent key: ", result.Error)
		return "", fmt.Errorf("error fetching idempotent key: %w", result.Error)
	}

	log.Infof("Idempotent key %s found, customer ID: %s", idempotentKey, Idempotent.CustomerID)

	if Idempotent.CustomerID == "" {
		log.Error("Customer ID is empty for idempotent key: ", idempotentKey)
		return "", fmt.Errorf("customer ID is empty for idempotent key: %s", idempotentKey)
	}

	// Fetch seat details from the database using bookedSeatsId

	// var bookedSeats []models.BookedSeats

	// for _, v := range Idempotent.BookedSeatsId {

	// 	var seat models.BookedSeats

	// 	result := c.DB.Model(&models.BookedSeats{}).Where("id = ?", v).First(&seat)

	// 	if result.Error != nil {
	// 		if result.Error == gorm.ErrRecordNotFound {
	// 			log.Errorf("Booked seat with ID %d not found", v)
	// 			return "", fmt.Errorf("booked seat with ID %d not found", v)
	// 		}
	// 		log.Error("Error fetching booked seat: ", result.Error)
	// 		return "", fmt.Errorf("error fetching booked seat: %w", result.Error)
	// 	}

	// 	log.Infof("Booked seat found: %s for movie time slot ID: %d", seat.SeatNumber, Idempotent.MovieTimeSlotID)

	// 	bookedSeats = append(bookedSeats, seat)
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var productCartArr []dodopayments.OneTimeProductCartItemParam

	for _, v := range Idempotent.OrderIDs {
		productCartArr = append(productCartArr, dodopayments.OneTimeProductCartItemParam{
			ProductID: dodopayments.F(v),
			Quantity:  dodopayments.F(int64(1)), // Assuming quantity is always 1 for each product
		})
	}

	dodopayments.NewWebhookEventService()

	paymentLink, err := c.Client.Payments.New(ctx, dodopayments.PaymentNewParams{
		Billing: dodopayments.F(dodopayments.BillingAddressParam{
			Country: dodopayments.F(dodopayments.CountryCodeIn),
			State:   dodopayments.F("Karnataka"),          // Replace with your state
			City:    dodopayments.F("Banglore"),           // Replace with your city
			Street:  dodopayments.F("123 Example Street"), // Replace with your street
			Zipcode: dodopayments.F("560001"),             // Replace with your zipcode
		}),
		Customer: dodopayments.F[dodopayments.CustomerRequestUnionParam](dodopayments.AttachExistingCustomerParam{
			CustomerID: dodopayments.F(Idempotent.CustomerID),
		}),
		ProductCart:     dodopayments.F(productCartArr),
		PaymentLink:     dodopayments.F(true),
		ReturnURL:       dodopayments.F("https://example.com/return"), // Replace with your return URL
		BillingCurrency: dodopayments.F(dodopayments.CurrencyInr),
		Metadata: dodopayments.F(map[string]string{
			"idempotent_key":     idempotentKey,
			"movie_time_slot_id": fmt.Sprint(Idempotent.MovieTimeSlotID),
			"booked_seats_id":    string(Idempotent.BookedSeatsId),
			"customer_id":        Idempotent.CustomerID,
		}),
	})

	if err != nil {
		log.Error("Failed to create payment link: ", err)
		return "", fmt.Errorf("failed to create payment link: %w", err)
	}

	log.Infof("Payment link created successfully: %s", paymentLink.PaymentLink)

	return paymentLink.PaymentLink, nil
}

func (m *Payment_Service) Update_Wallet_Ledger(idempotent_key string, transaction_id string) error {

	tx := m.DB.Begin()

	if tx.Error != nil {
		log.Error("Failed to begin transaction: ", tx.Error)
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("Transaction rolled back due to panic: ", r)
		}
	}()

	// Fetch the payment details using the transaction id

	url := fmt.Sprintf("https://test.dodopayments.com/payments/%s", transaction_id)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		tx.Rollback()
		log.Error("Failed to create HTTP request: ", err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", os.Getenv("DODOPAYMENT_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		tx.Rollback()
		log.Error("Failed to send HTTP request: ", err)
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tx.Rollback()
		log.Errorf("Failed to fetch payment details, status code: %d", resp.StatusCode)
		return fmt.Errorf("failed to fetch payment details, status code: %d", resp.StatusCode)
	}

	var paymentDetail PaymentDetail

	err = json.NewDecoder(resp.Body).Decode(&paymentDetail)

	if err != nil {
		tx.Rollback()
		log.Error("Failed to decode payment details: ", err)
		return fmt.Errorf("failed to decode payment details: %w", err)
	}

	log.Infof("Payment details fetched successfully for transaction ID: %s", transaction_id)

	// Use the payment object to fetch customer details and product details and update wallet and ledger accordingly

	customerID := paymentDetail.Customer.CustomerID

	orderIds := make([]string, len(paymentDetail.ProductCart))

	for i, product := range paymentDetail.ProductCart {
		orderIds[i] = product.ProductID
	}

	url = fmt.Sprintf("https://test.dodopayments.com/customers/%s", customerID)

	req, err = http.NewRequest("GET", url, nil)

	if err != nil {
		tx.Rollback()
		log.Error("Failed to create HTTP request for customer details: ", err)
		return fmt.Errorf("failed to create HTTP request for customer details: %w", err)
	}

	req.Header.Set("Authorization", os.Getenv("DODOPAYMENT_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	if err != nil {
		tx.Rollback()
		log.Error("Failed to send HTTP request for customer details: ", err)
		return fmt.Errorf("failed to send HTTP request for customer details: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tx.Rollback()
		log.Errorf("Failed to fetch customer details, status code: %d", resp.StatusCode)
		return fmt.Errorf("failed to fetch customer details, status code: %d", resp.StatusCode)
	}

	var customerDetail CustomerDetail

	err = json.NewDecoder(resp.Body).Decode(&customerDetail)

	if err != nil {
		tx.Rollback()
		log.Error("Failed to decode customer details: ", err)
		return fmt.Errorf("failed to decode customer details: %w", err)
	}

	log.Infof("Customer details fetched successfully for customer ID: %s", customerID)

	// Get product details from the database using order IDs

	var products []ProductDetail

	for _, orderID := range orderIds {
		url = fmt.Sprintf("https://test.dodopayments.com/products/%s", orderID)

		req, err = http.NewRequest("GET", url, nil)

		if err != nil {
			tx.Rollback()
			log.Error("Failed to create HTTP request for product details: ", err)
			return fmt.Errorf("failed to create HTTP request for product details: %w", err)
		}

		req.Header.Set("Authorization", os.Getenv("DODOPAYMENT_TOKEN"))
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)

		if err != nil {
			tx.Rollback()
			log.Error("Failed to send HTTP request for product details: ", err)
			return fmt.Errorf("failed to send HTTP request for product details: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tx.Rollback()
			log.Errorf("Failed to fetch product details for order ID %s, status code: %d", orderID, resp.StatusCode)
			return fmt.Errorf("failed to fetch product details for order ID %s, status code: %d", orderID, resp.StatusCode)
		}

		var productDetail ProductDetail

		err = json.NewDecoder(resp.Body).Decode(&productDetail)

		if err != nil {
			tx.Rollback()
			log.Error("Failed to decode product details: ", err)
			return fmt.Errorf("failed to decode product details: %w", err)
		}

		log.Infof("Product details fetched successfully for product ID: %s", productDetail.ProductID)

		products = append(products, productDetail)

	}

	// Update the wallet balance and create a ledger entry

	wallet := models.Wallet{
		UserID:   customerDetail.CustomerID,
		Balance:  float64(paymentDetail.TotalAmount) / 100, // Convert to float64
		Currency: paymentDetail.Currency,
	}

	result := tx.Model(&models.Wallet{}).Create(&wallet)

	if result.Error != nil {
		tx.Rollback()
		log.Error("Failed to create wallet entry: ", result.Error)
		return fmt.Errorf("failed to create wallet entry: %w", result.Error)
	}

	log.Infof("Wallet entry created successfully for user ID: %s", wallet.UserID)

	var ledgerArr []models.Ledger

	for _, product := range products {
		ledger := models.Ledger{
			WalletID:      wallet.ID,
			OrderID:       nil, // Assuming no order ID is associated
			TransactionID: transaction_id,
			Amount:        float64(product.Price.Price) / 100, // Convert to float64
			Type:          "credit",
			Description:   fmt.Sprintf("Payment received for product %s", product.Name),
			PSPRefID:      paymentDetail.PaymentID,
		}

		ledgerArr = append(ledgerArr, ledger)
	}

	for _, ledger := range ledgerArr {
		result := tx.Model(&models.Ledger{}).Create(&ledger)

		if result.Error != nil {
			tx.Rollback()
			log.Error("Failed to create ledger entry: ", result.Error)
			return fmt.Errorf("failed to create ledger entry: %w", result.Error)
		}

		log.Infof("Ledger entry created successfully for wallet ID: %d", ledger.WalletID)
	}

	if err := tx.Commit().Error; err != nil {
		log.Error("Failed to commit transaction: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info("Transaction committed successfully")

	return nil
}
