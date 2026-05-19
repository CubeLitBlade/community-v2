// Package setup provides constructors that wire together account domain services,
// persistence, and HTTP transport.
package setup

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/postgres"
	"github.com/CubeLitBlade/community-v2/backend/internal/account/transport"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

func NewReader(db *gorm.DB) *postgres.Reader {
	return postgres.NewReader(db)
}

func NewWriter(db *gorm.DB) *postgres.Writer {
	return postgres.NewWriter(db)
}

// NewHandler creates the account HTTP handler with all dependencies wired.
func NewHandler(writer *postgres.Writer, ids idgen.Generator, logger *slog.Logger) *transport.Handler {
	registrar := account.NewRegistrar(ids, writer, logger)

	return transport.NewHandler(transport.Deps{
		Registrar: registrar,
		Logger:    logger,
	})
}

// NewAuthenticator creates the account authenticator for cross-module injection.
func NewAuthenticator(reader *postgres.Reader, logger *slog.Logger) *account.Authenticator {
	return account.NewAuthenticator(reader, logger)
}

func NewLoginRecorder(writer *postgres.Writer, logger *slog.Logger) *account.LoginRecorder {
	return account.NewLoginRecorder(time.Now, writer, logger)
}
