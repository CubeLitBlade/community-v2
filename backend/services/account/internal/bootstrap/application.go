package bootstrap

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/pkg/common/idgen"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driven/hasher"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/adapter/driven/persistence"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/domain/port"
	"github.com/cubelitblade/community-v2/backend/services/account/internal/telemetry"
)

func ApplicationModule() fx.Option {
	return fx.Options(
		fx.Provide(NewIDGeneratorAdapter),
		fx.Provide(NewPasswordHasherAdapter),
		fx.Provide(persistence.NewAccountRepository),

		fx.Provide(func(
			repository *persistence.AccountRepository,
			idGen port.IDGenerator,
			tel *telemetry.Telemetry,
			logger *slog.Logger,
		) application.RegistrarUseCase {
			inner := application.NewRegistrar(idGen, repository)
			return telemetry.NewInstrumentedRegistrar(inner, tel, logger)
		}),

		fx.Provide(func(
			repository *persistence.AccountRepository,
			tel *telemetry.Telemetry,
			logger *slog.Logger,
		) application.AuthenticatorUseCase {
			inner := application.NewAuthenticator(repository)
			return telemetry.NewInstrumentedAuthenticator(inner, tel, logger)
		}),

		fx.Provide(func(
			repository *persistence.AccountRepository,
			tel *telemetry.Telemetry,
			logger *slog.Logger,
		) application.ProfileFinderUseCase {
			inner := application.NewProfileFinder(repository)
			return telemetry.NewInstrumentedProfileFinder(inner, tel, logger)
		}),
	)
}

func NewIDGeneratorAdapter(gen idgen.Generator) port.IDGenerator {
	return &idGeneratorAdapter{gen: gen}
}

type idGeneratorAdapter struct {
	gen idgen.Generator
}

func (a *idGeneratorAdapter) NextID() (int64, error) {
	return a.gen.NextID()
}

func NewPasswordHasherAdapter() port.PasswordHasher {
	return hasher.NewBcryptHasher()
}
