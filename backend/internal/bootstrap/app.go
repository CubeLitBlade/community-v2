package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	accountSetup "github.com/CubeLitBlade/community-v2/backend/internal/account/setup"
	"github.com/CubeLitBlade/community-v2/backend/internal/auth"
	authSetup "github.com/CubeLitBlade/community-v2/backend/internal/auth/setup"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/jwt"
)

const defaultJWTValidity = 24 * time.Hour

// NewApp creates the fx application with all providers and lifecycle hooks.
func NewApp() *fx.App {
	return fx.New(
		infraProviders(),
		moduleProviders(),
		fx.Invoke(registerRoutes),
		fx.Invoke(func(*http.Server) {}),
	)
}

// --- Infrastructure providers ---

func infraProviders() fx.Option {
	return fx.Options(
		fx.Provide(LoadConfig),
		fx.Provide(NewAppLogger),
		fx.Provide(provideDatabase),
		fx.Provide(
			fx.Annotate(provideSnowflake, fx.As(new(idgen.Generator))),
		),
		fx.Provide(newRouter),
	)
}

func provideDatabase(lifecycle fx.Lifecycle, cfg Config) (db *gorm.DB, sqlDB *sql.DB, err error) {
	db, sqlDB, err = OpenDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, sqlDB, nil
}

func provideSnowflake(cfg Config) (*idgen.Snowflake, error) {
	s, err := idgen.NewSnowflake(cfg.SnowflakeID)
	if err != nil {
		return nil, fmt.Errorf("create snowflake: %w", err)
	}

	return s, nil
}

// --- Module providers ---

func moduleProviders() fx.Option {
	return fx.Options(
		fx.Provide(provideJWTConfig),
		fx.Provide(accountSetup.NewReader),
		fx.Provide(accountSetup.NewWriter),
		fx.Provide(
			fx.Annotate(
				accountSetup.NewHandler,
				fx.As(new(HTTPMounter)),
				fx.ResultTags(`group:"mounter"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				accountSetup.NewAuthenticator,
				fx.As(new(auth.AccountAuthenticator)),
			),
		),
		fx.Provide(
			fx.Annotate(
				accountSetup.NewLoginRecorder,
				fx.As(new(auth.LoginRecorder)),
			),
		),
		fx.Provide(
			fx.Annotate(
				authSetup.NewHandler,
				fx.As(new(HTTPMounter)),
				fx.ResultTags(`group:"mounter"`),
			),
		),
		fx.Provide(provideHTTPServer),
	)
}

func provideJWTConfig(cfg Config) *jwt.Config {
	return &jwt.Config{
		Key:      cfg.JWTSecret,
		Issuer:   "community-v2",
		Validity: defaultJWTValidity,
	}
}

func provideHTTPServer(lc fx.Lifecycle, cfg Config, router *gin.Engine) *http.Server {
	srv := newHTTPServer(&cfg, router)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
