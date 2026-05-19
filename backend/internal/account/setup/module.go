// Package setup provides constructors that wire together account domain services,
// persistence, and HTTP transport.
package setup

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/transport"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

// NewHandler creates the account HTTP handler with all dependencies wired.
func NewHandler(db *gorm.DB, ids idgen.Generator, logger *slog.Logger) *transport.Handler {
	writer := postgres.NewWriter(db)

	registrar := account.NewRegistrar(ids, writer, logger)

	return transport.NewHandler(transport.Deps{
		Registrar: registrar,
		Logger:    logger,
	})
}

// NewAuthenticator creates the account authenticator for cross-module injection.
func NewAuthenticator(db *gorm.DB, logger *slog.Logger) *account.Authenticator {
	reader := postgres.NewReader(db)

	return account.NewAuthenticator(reader, logger)
}
