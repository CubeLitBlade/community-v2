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
	authTransport "github.com/CubeLitBlade/community-v2/backend/internal/auth/transport"
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
		// Infrastructure
		fx.Provide(provideJWTConfig),
		fx.Provide(provideCookieConfig),

		// Account module
		accountSetup.Module(),
		fx.Provide(
			fx.Annotate(
				accountSetup.NewHandler,
				fx.As(new(HTTPMounter)),
				fx.ResultTags(`group:"mounter"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				accountSetup.NewAuthenticatorAdapter,
				fx.As(new(auth.AccountAuthenticator)),
			),
		),
		fx.Provide(
			fx.Annotate(
				accountSetup.NewLoginRecorder,
				fx.As(new(auth.LoginRecorder)),
			),
		),

		// Auth module
		authSetup.Module(),
		fx.Provide(
			fx.Annotate(
				authSetup.NewHandler,
				fx.As(new(HTTPMounter)),
				fx.ResultTags(`group:"mounter"`),
			),
		),

		// Server
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

func provideCookieConfig(cfg Config) authTransport.CookieConfig {
	return authTransport.CookieConfig{
		Name:     cfg.AccessTokenCookieName,
		Secure:   cfg.CookieSecure,
		SameSite: cfg.CookieSameSite,
		MaxAge:   int(cfg.AccessTokenTTL.Seconds()),
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
