package port

// PaymentEvent es el payload canónico para eventos de dominio de métodos de pago (ADR-001).
// Campos flat, named. Los nombres comunes (tenant_id, user_id) son idénticos al resto
// de la flota para que el LogQL cross-service funcione. Todos opcionales salvo Event.
//
// CRÍTICO (servicio de dinero/pagos): jamás incluir números de tarjeta, tokens,
// datos de pago sensibles. Solo IDs/códigos no sensibles.
type PaymentEvent struct {
	Event            string // <domain>.<action>_<result>, p.ej. "payment.method_created"
	TenantID         string
	UserID           string
	PaymentMethodID  string
	PaymentMethodCode string
	Reason           string
}

// PaymentEventLogger es el puerto para emitir eventos canónicos de métodos de pago.
// El código de aplicación depende de esta interfaz; el adapter (JSON a stdout,
// Loki push, etc.) la implementa. Nunca al revés.
type PaymentEventLogger interface {
	Log(e PaymentEvent)
}
