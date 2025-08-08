package test

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/kartik7120/booking_payment_service/cmd/api/models"
	"github.com/kartik7120/booking_payment_service/cmd/api/server"
)

func TestMigrateDB(t *testing.T) {

	t.Run("MigrateDB", func(t *testing.T) {

		err := godotenv.Load()

		if err != nil {
			t.Fatalf("Error loading .env file: %v", err)
		}

		m := server.NewPaymentServer()

		m.Ps.DB.AutoMigrate(&models.Idempotent{})

		// m.Ps.DB.AutoMigrate(&models.Payment{})
		// m.Ps.DB.AutoMigrate(&models.Order{})
	})

	t.Run("DropTables", func(t *testing.T) {
		err := godotenv.Load()

		if err != nil {
			t.Fatalf("Error loading .env file: %v", err)
		}

		m := server.NewPaymentServer()

		m.Ps.DB.Migrator().DropTable(&models.Idempotent{})
	})
}
