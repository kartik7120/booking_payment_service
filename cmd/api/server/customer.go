package server

import (
	"context"
	"errors"

	"github.com/dodopayments/dodopayments-go"
	log "github.com/sirupsen/logrus"
)

func (m *Payment_Service) Create_Customer(email string, name string, phone_number string) (*dodopayments.Customer, error) {

	// Validate the input parameters

	if email == "" || name == "" || phone_number == "" {
		log.Error("Email, name, and phone number must be provided")
		return nil, errors.New("email, name, and phone number must be provided")
	}

	if m.Validator.Var(email, "email") != nil {
		log.Error("Invalid email format: ", email)
		return nil, errors.New("invalid email format")
	}

	if m.Validator.Var(phone_number, "e164") != nil {
		log.Error("Invalid phone number format: ", phone_number)
		return nil, errors.New("invalid phone number format")
	}

	if m.Validator.Var(name, "required") != nil {
		log.Error("Name is required")
		return nil, errors.New("name is required")
	}

	customer, err := m.Client.Customers.New(context.Background(), dodopayments.CustomerNewParams{
		Email:       dodopayments.F(email),
		PhoneNumber: dodopayments.F(phone_number), // Replace with actual phone number
		Name:        dodopayments.F(name),         // Replace with actual customer name
	})

	if err != nil {
		log.Error("Failed to create customer: ", err)
		return nil, err
	}

	log.Info("Customer created successfully with ID: ", customer.CustomerID)

	return customer, nil
}
