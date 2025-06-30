package models

import (
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
