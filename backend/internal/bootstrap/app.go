package bootstrap

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	cfg    Config
	logger *slog.Logger
	db     *gorm.DB
	sqlDB  *sql.DB
	router *gin.Engine
	server *http.Server
}

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
		_ = sqlDB.Close()
		return nil, fmt.Errorf("create id generator: %w", err)
	}

	// build router and register modules
	router, err := newRouter()
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	RegisterModules(router, ModuleDeps{
		DB:     db,
		IDs:    ids,
		Logger: appLogger,
	})

	// build HTTP server
	server, err := newHTTPServer(cfg, router)
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &App{
		cfg:    cfg,
		logger: appLogger,
		db:     db,
		sqlDB:  sqlDB,
		router: router,
		server: server,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("server started", "addr", a.cfg.Addr)

	if err := a.server.ListenAndServe(); err != nil {
		a.logger.Error("server stopped unexpectedly", "error", err)
		return err
	}

	return nil
}

func (a *App) Close() {
	if a.sqlDB == nil {
		return
	}

	if err := a.sqlDB.Close(); err != nil {
		a.logger.Error("close database failed", "error", err)
	}
}
