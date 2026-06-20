package payment_method

import "embed"

// MigrationsFS embeds all migration files for payment-method-service.
// The "migrations" subdirectory name is required by the go-shared migrate helper
// (iofs.New expects the files under a named subdirectory of the provided FS).
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
