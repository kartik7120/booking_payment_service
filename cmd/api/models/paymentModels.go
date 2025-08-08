package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	MovieID         uint   `json:"movie_id" gorm:"not null"`              // ID of the movie for which the payment is made
	Amount          uint   `json:"amount" gorm:"not null"`                // Amount to be paid
	Email           string `json:"email" gorm:"not null"`                 // Email of the user making the payment
	Phone           string `json:"phone" gorm:"not null"`                 // Phone number of the user making the payment
	Address         string `json:"address" gorm:"not null"`               // Address of the user making the payment
	MovieName       string `json:"movie_name" gorm:"not null"`            // Name of the movie for which the payment is made
	PaymentStatus   string `json:"payment_status" gorm:"not null"`        // Status of the payment (e.g., pending, completed, failed)
	PaymentMethod   string `json:"payment_method" gorm:"not null"`        // Method of payment (e.g., credit card, PayPal)
	TransactionID   string `json:"transaction_id" gorm:"not null;unique"` // Unique transaction ID for the payment
	Quantity        uint   `json:"quantity" gorm:"not null"`              // Number of tickets purchased
	Price           uint   `json:"price" gorm:"not null"`                 // Price per ticket
	CustomerID      string `json:"customer_id" gorm:"not null"`           // Unique ID of the customer making the payment
	VenueID         uint   `json:"venue_id" gorm:"not null"`              // ID of the venue where the movie is being shown
	MovieTimeSlotID uint   `json:"movie_time_slot_id" gorm:"not null"`    // ID of the movie time slot for which the payment is made
}

type Order struct {
	gorm.Model
	ProductID  uint   `json:"product_id" gorm:"not null"`  // ID of the product being ordered
	Quantity   uint   `json:"quantity" gorm:"not null"`    // Number of products ordered
	Price      uint   `json:"price" gorm:"not null"`       // Price per product
	PaymentID  uint   `json:"payment_id" gorm:"not null"`  // ID of the payment associated with the order
	CustomerID string `json:"customer_id" gorm:"not null"` // Unique ID of the customer placing the order
}

type Idempotent struct {
	gorm.Model
	PaymentID     string         `json:"payment_id" gorm:"not null;unique"`     // Unique Idempotency key for the payment
	CustomerID    string         `json:"customer_id" gorm:"not null"`           // Unique ID of the customer associated with the idempotency key
	IdempotentKey string         `json:"idempotent_key" gorm:"not null;unique"` // Unique idempotency key to ensure the operation is not repeated
	OrderIDs      pq.StringArray `json:"order_ids" gorm:"type:text[]"`          // List of order IDs associated with the idempotency key
	// CreatedAt  int64  `json:"created_at" gorm:"not null"`        // Timestamp when the idempotency key was created
	// UpdatedAt  int64  `json:"updated_at" gorm:"not null"`        // Timestamp when the idempotency key was last updated
	// DeletedAt  *int64 `json:"deleted_at" gorm:"index"`           // Timestamp when the idempotency key was deleted, if applicable
	// ID         uint   `json:"id" gorm:"primaryKey"`              // Primary key for the idempotency record
	ExpiredAt     time.Time `json:"expired_at" gorm:"not null"`     // Timestamp when the idempotency key expires
	PaymentStatus string    `json:"payment_status" gorm:"not null"` // Status of the payment associated with the idempotency key
	// VenueID         uint          `json:"venue_id" gorm:"not null"`       // ID of the venue associated with the idempotency key
	// MovieID         uint          `json:"movie_id" gorm:"not null"`
	BookedSeatsId   pq.Int32Array `json:"booked_seats_id" gorm:"type:integer[]"` // List of booked seat IDs associated with the idempotency key
	MovieTimeSlotID uint          `json:"movie_time_slot_id" gorm:"not null"`    // ID of the movie time slot associated with the idempotency key
	IsTicketSent    bool          `json:"is_ticket_sent" gorm:"not null"`        // Flag to indicate if the ticket has been sent
	IsMailSend      bool          `json:"is_mail_send" gorm:"not null"`          // Flag to indicate if the mail has been sent
}

type BookedSeats struct {
	gorm.Model
	SeatNumber      string  `json:"seat_number" gorm:"not null;uniqueIndex:idx_unique_booked_seats"`
	MovieTimeSlotID uint    `json:"movie_time_slot_id" gorm:"not null;uniqueIndex:idx_unique_booked_seats"` // Link booking to a movie show
	SeatMatrixID    uint    `json:"seat_matrix_id" gorm:"not null;uniqueIndex:idx_unique_booked_seats"`     // Reference seat matrix for consistency
	IsBooked        bool    `json:"is_booked"`
	Email           *string `json:"email" validate:"required,email"`
	PhoneNumber     string  `json:"phone_number" validate:"required,e164"`
}

// type Wallet struct {
// 	gorm.Model
// 	CustomerID    string `json:"customer_id" gorm:"not null;unique"`    //
// 	Balance       uint   `json:"balance" gorm:"not null"`               // Current balance in the wallet
// 	TransactionID string `json:"transaction_id" gorm:"not null;unique"` // Unique transaction ID for the wallet operation
// 	Amount        uint   `json:"amount" gorm:"not null"`                // Amount to be added or deducted from the walle
// 	// TransactionType indicates whether the transaction is a credit or debit
// 	// TransactionType string `json:"transaction_type" gorm:"not null"` // Type of
// 	TransactionType string `json:"transaction_type" gorm:"not null"`     // Type of transaction (credit or debit)
// 	IdempotentID    string `json:"idempotent_id" gorm:"not null;unique"` // Unique idempotency key to ensure the operation is not repeated
// }

// Wallet model to manage user wallets
// It tracks the balance, transactions, and associated user
// It can be used for payments, refunds, and other financial operations
type Wallet struct {
	gorm.Model
	UserID   string  `gorm:"index"` // Associated user, a phone number
	Balance  float64 `gorm:"type:numeric(12,2)"`
	Currency string  `gorm:"size:3;default:'INR'"`
}

// type Ledger struct {
// 	gorm.Model
// 	CustomerID      string `json:"customer_id" gorm:"not null"`           //
// 	TransactionID   string `json:"transaction_id" gorm:"not null;unique"` // Unique transaction ID for the ledger entry
// 	Amount          uint   `json:"amount" gorm:"not null"`                // Amount involved in the transaction
// 	TransactionType string `json:"transaction_type" gorm:"not null"`      // Type
// 	// TransactionType string `json:"transaction_type" gorm:"not null"` // Type of transaction (credit or debit)
// 	PaymentID string `json:"payment_id" gorm:"not null"` //
// 	// PaymentID     string `json:"payment_id" gorm:"not null"`            // ID of the payment associated with the ledger entry
// 	WalletID string `json:"wallet_id" gorm:"not null"`
// 	// ID of the wallet associated with the ledger entry
// }

// Ledger is a record of transactions in the wallet
// It tracks credits, debits, and refunds
// It can be used for auditing and reporting purposes
type Ledger struct {
	gorm.Model
	WalletID      uint    `gorm:"index"`       // Wallet association
	OrderID       *uint   `gorm:"index"`       // Optional: ties to order
	TransactionID string  `gorm:"uniqueIndex"` // External PSP reference
	Amount        float64 `gorm:"type:numeric(12,2)"`
	Type          string  `gorm:"size:20"` // 'credit', 'debit', 'refund'
	Description   string  `gorm:"size:255"`
	PSPRefID      string  `gorm:"size:100"` // Optional external ref
}
