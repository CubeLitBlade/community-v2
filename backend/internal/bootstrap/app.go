package bootstrap

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
)

// App holds the core dependencies and infrastructure for the application.
type App struct {
	logger *slog.Logger
	db     *gorm.DB
	sqlDB  *sql.DB
	router *gin.Engine
	server *http.Server
	cfg    Config
}

// NewApp initializes configuration, infrastructure, and HTTP server,
// returning a ready-to-run App.
func NewApp() (*App, error) {
	// load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// initialize infrastructure
	appLogger := NewAppLogger()

	db, sqlDB, err := OpenDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	ids, err := idgen.NewSnowflake(cfg.SnowflakeID)
	if err != nil {
		return nil, closeSQLDBAfterError(
			sqlDB,
			fmt.Errorf("create id generator: %w", err),
		)
	}

	// build router and register modules
	router, err := newRouter()
	if err != nil {
		return nil, closeSQLDBAfterError(
			sqlDB,
			fmt.Errorf("create router: %w", err),
		)
	}

	RegisterModules(router, ModuleDeps{
		DB:     db,
		IDs:    ids,
		Logger: appLogger,
	})

	// build HTTP server
	server := newHTTPServer(&cfg, router)

	return &App{
		cfg:    cfg,
		logger: appLogger,
		db:     db,
		sqlDB:  sqlDB,
		router: router,
		server: server,
	}, nil
}

// Run starts the HTTP server to listen for incoming requests.
func (a *App) Run() error {
	a.logger.Info("server started", "addr", a.cfg.Addr)

	err := a.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

// Close releases resources held by the App, including closing the database
// connection.
func (a *App) Close() {
	if a.sqlDB == nil {
		return
	}

	err := a.sqlDB.Close()
	if err != nil {
		a.logger.Error("close database failed", "error", err)
	}
}

func closeSQLDBAfterError(db *sql.DB, cause error) error {
	if db == nil {
		return cause
	}

	closeErr := db.Close()
	if closeErr != nil {
		return errors.Join(cause, fmt.Errorf("close database: %w", closeErr))
	}

	return cause
}
