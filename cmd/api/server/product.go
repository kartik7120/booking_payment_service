package server

import (
	"context"
	"fmt"

	"github.com/dodopayments/dodopayments-go"
	log "github.com/sirupsen/logrus"
)

type Product struct {
	ProductName        string `json:"product_name" validate:"required"`
	Price              int64  `json:"price" validate:"required;min=1"`
	ProductDescription string `json:"product_description" validate:"required"`
}

func (m *Payment_Service) Create_Product_Ticket(product Product) (*dodopayments.Product, error) {
	// Create a product with the given name and price
	p, err := m.Client.Products.New(context.Background(), dodopayments.ProductNewParams{
		Price: dodopayments.F[dodopayments.PriceUnionParam](dodopayments.PriceOneTimePriceParam{
			Currency: dodopayments.F(dodopayments.CurrencyInr),
			Price:    dodopayments.F(product.Price),
		}),
		Name:        dodopayments.F(product.ProductName),
		Description: dodopayments.F(product.ProductDescription),
	})

	if err != nil {
		log.Error("Failed to create product: ", err)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	log.Info("Product created successfully: ", p.ProductID)

	return p, nil
}
