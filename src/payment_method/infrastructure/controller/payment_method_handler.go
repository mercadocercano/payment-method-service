package controller

import (
	"net/http"
	"payment_method/src/payment_method/application/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hornosg/go-shared/infrastructure/response"
)

// PaymentMethodHandler maneja las peticiones HTTP relacionadas con métodos de pago
type PaymentMethodHandler struct {
	getByIDUseCase *usecase.GetPaymentMethodByIDUseCase
	listUseCase    *usecase.ListPaymentMethodsUseCase
}

// NewPaymentMethodHandler crea una nueva instancia del handler
func NewPaymentMethodHandler(
	getByIDUseCase *usecase.GetPaymentMethodByIDUseCase,
	listUseCase *usecase.ListPaymentMethodsUseCase,
) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		getByIDUseCase: getByIDUseCase,
		listUseCase:    listUseCase,
	}
}

// GetByID maneja GET /payment-methods/:id
func (h *PaymentMethodHandler) GetByID(c *gin.Context) {
	// Extraer tenant ID del header
	tenantIDStr := c.GetHeader("X-Tenant-ID")
	if tenantIDStr == "" {
		response.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}

	// Extraer payment method ID del path
	paymentMethodIDStr := c.Param("id")
	paymentMethodID, err := uuid.Parse(paymentMethodIDStr)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, "invalid payment method id format")
		return
	}

	// Ejecutar caso de uso
	paymentMethod, err := h.getByIDUseCase.Execute(paymentMethodID, tenantID)
	if err != nil {
		if err.Error() == "payment method not found" {
			response.JSON(c, http.StatusNotFound, "payment method not found")
			return
		}
		response.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, paymentMethod)
}

// List maneja GET /payment-methods
func (h *PaymentMethodHandler) List(c *gin.Context) {
	// Extraer tenant ID del header
	tenantIDStr := c.GetHeader("X-Tenant-ID")
	if tenantIDStr == "" {
		response.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.JSON(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}

	// Extraer parámetro de query
	activeOnlyStr := c.DefaultQuery("active_only", "true")
	activeOnly, _ := strconv.ParseBool(activeOnlyStr)

	// Ejecutar caso de uso
	result, err := h.listUseCase.Execute(tenantID, activeOnly)
	if err != nil {
		response.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}
