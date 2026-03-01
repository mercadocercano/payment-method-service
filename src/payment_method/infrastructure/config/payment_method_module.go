package config

import (
	"database/sql"
	"payment_method/src/payment_method/application/usecase"
	"payment_method/src/payment_method/infrastructure/controller"
	"payment_method/src/payment_method/infrastructure/persistence/repository"

	"github.com/gin-gonic/gin"
)

// SetupPaymentMethodModule configura el módulo de métodos de pago
func SetupPaymentMethodModule(router *gin.RouterGroup, db *sql.DB) {
	// Repository
	paymentMethodRepository := repository.NewPostgresPaymentMethodRepository(db)

	// Use Cases
	getByIDUseCase := usecase.NewGetPaymentMethodByIDUseCase(paymentMethodRepository)
	listUseCase := usecase.NewListPaymentMethodsUseCase(paymentMethodRepository)

	// Handler
	handler := controller.NewPaymentMethodHandler(getByIDUseCase, listUseCase)

	// Routes
	paymentMethods := router.Group("/payment-methods")
	{
		paymentMethods.GET("", handler.List)
		paymentMethods.GET("/:id", handler.GetByID)
	}
}
