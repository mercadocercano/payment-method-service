package logging

import (
	"io"

	"payment_method/src/payment_method/domain/port"

	sharedlog "github.com/hornosg/go-shared/infrastructure/logging"
)

// PaymentLogger implementa port.PaymentEventLogger emitiendo una línea JSON canónica
// (ADR-001) por evento, delegando el envelope (ts/level/service/event + campos flat
// omitempty) en go-shared CanonicalLogger (>= v0.8.0). El mapeo struct→fields y las
// reglas de nivel por evento viven acá; el formato canónico es compartido por la flota.
//
// CRÍTICO: este servicio maneja datos de pagos. Jamás loguear números de tarjeta,
// tokens, secrets ni ningún dato sensible — solo IDs/códigos no sensibles.
type PaymentLogger struct {
	canonical *sharedlog.CanonicalLogger
}

// NewPaymentLogger crea el adapter escribiendo a stdout. El service se fija acá, nunca por-call.
func NewPaymentLogger(service string) *PaymentLogger {
	return &PaymentLogger{canonical: sharedlog.NewCanonicalLogger(service)}
}

// NewPaymentLoggerWithWriter permite inyectar un io.Writer (tests).
func NewPaymentLoggerWithWriter(service string, w io.Writer) *PaymentLogger {
	return &PaymentLogger{canonical: sharedlog.NewCanonicalLoggerWithWriter(service, w)}
}

// levelFor aplica las reglas de nivel del ADR-001 por tipo de evento.
func levelFor(event string) string {
	switch event {
	case "payment.method_found":
		return "info"
	case "payment.methods_listed":
		return "info"
	case "payment.method_not_found":
		return "warn"
	case "payment.method_fetch_failed", "payment.methods_list_failed":
		return "error"
	default:
		return "info"
	}
}

// Log traduce el struct tipado a campos flat y delega en CanonicalLogger.
func (l *PaymentLogger) Log(e port.PaymentEvent) {
	fields := map[string]any{
		"tenant_id":          e.TenantID,
		"user_id":            e.UserID,
		"payment_method_id":  e.PaymentMethodID,
		"payment_method_code": e.PaymentMethodCode,
		"reason":             e.Reason,
	}
	l.canonical.Emit(levelFor(e.Event), e.Event, fields)
}
