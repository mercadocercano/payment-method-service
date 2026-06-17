package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"payment_method/src/payment_method/domain/port"
	paymentlog "payment_method/src/payment_method/infrastructure/logging"

	"github.com/stretchr/testify/assert"
)

// ADR-001: cada evento produce UNA línea JSON canónica con envelope ts/level/service/event.
func parseLine(t *testing.T, b []byte) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(b), []byte("\n"))
	assert.Len(t, lines, 1, "debe ser exactamente una línea por evento")
	var m map[string]any
	assert.NoError(t, json.Unmarshal(lines[0], &m))
	return m
}

func TestPaymentLogger_MethodFound_EnvelopeAndInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := paymentlog.NewPaymentLoggerWithWriter("payment-test", &buf)

	logger.Log(port.PaymentEvent{
		Event:             "payment.method_found",
		TenantID:          "t-123",
		PaymentMethodID:   "pm-456",
		PaymentMethodCode: "cash",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "payment.method_found", line["event"])
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "payment-test", line["service"])
	assert.NotEmpty(t, line["ts"], "ts (RFC3339 UTC) siempre presente")
	assert.Equal(t, "t-123", line["tenant_id"])
	assert.Equal(t, "pm-456", line["payment_method_id"])
	assert.Equal(t, "cash", line["payment_method_code"])
}

func TestPaymentLogger_MethodNotFound_WarnLevel_OmitsEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := paymentlog.NewPaymentLoggerWithWriter("payment-test", &buf)

	logger.Log(port.PaymentEvent{
		Event:           "payment.method_not_found",
		TenantID:        "t-123",
		PaymentMethodID: "pm-999",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "payment.method_not_found", line["event"])
	assert.Equal(t, "pm-999", line["payment_method_id"])
	// omitempty: campos vacíos no aparecen
	_, hasCode := line["payment_method_code"]
	assert.False(t, hasCode, "payment_method_code vacío debe omitirse")
	_, hasReason := line["reason"]
	assert.False(t, hasReason, "reason vacío debe omitirse")
	// CRÍTICO: jamás campos sensibles de pago
	_, hasCard := line["card_number"]
	assert.False(t, hasCard, "nunca debe haber card_number")
	_, hasToken := line["token"]
	assert.False(t, hasToken, "nunca debe haber token")
}

func TestPaymentLogger_FetchFailed_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := paymentlog.NewPaymentLoggerWithWriter("payment-test", &buf)

	logger.Log(port.PaymentEvent{
		Event:           "payment.method_fetch_failed",
		TenantID:        "t-1",
		PaymentMethodID: "pm-1",
		Reason:          "db connection error",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "error", line["level"])
	assert.Equal(t, "payment.method_fetch_failed", line["event"])
	assert.Equal(t, "db connection error", line["reason"])
}

func TestPaymentLogger_MethodsListed_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := paymentlog.NewPaymentLoggerWithWriter("payment-test", &buf)

	logger.Log(port.PaymentEvent{
		Event:    "payment.methods_listed",
		TenantID: "t-123",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "payment.methods_listed", line["event"])
	assert.Equal(t, "t-123", line["tenant_id"])
}

func TestPaymentLogger_ListFailed_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := paymentlog.NewPaymentLoggerWithWriter("payment-test", &buf)

	logger.Log(port.PaymentEvent{
		Event:    "payment.methods_list_failed",
		TenantID: "t-1",
		Reason:   "replica lag",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "error", line["level"])
	assert.Equal(t, "payment.methods_list_failed", line["event"])
	assert.Equal(t, "replica lag", line["reason"])
}
