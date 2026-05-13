package bootstrap

import (
	"database/sql"
	"fmt"
	"log/slog"

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
}

func NewApp() (*App, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	appLogger := NewAppLogger()

	db, sqlDB, err := OpenDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	appLogger.Info("database connected")

	ids, err := idgen.NewSnowflake(cfg.SnowflakeID)
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("create id generator: %w", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if err := router.SetTrustedProxies(nil); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	RegisterModules(router, ModuleDeps{
		DB:     db,
		IDs:    ids,
		Logger: appLogger,
	})

	return &App{
		cfg:    cfg,
		logger: appLogger,
		db:     db,
		sqlDB:  sqlDB,
		router: router,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("server started", "addr", a.cfg.Addr)

	if err := a.router.Run(a.cfg.Addr); err != nil {
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
