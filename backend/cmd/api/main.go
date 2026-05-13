package main

import (
	"log/slog"
	"os"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/idgen"
	"github.com/CubeLitBlade/community-v2/backend/internal/web"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ids, err := idgen.NewSnowflake(1)
	if err != nil {
		logger.Error("create id generator failed", "error", err)
		os.Exit(1)
	}

	accountService := account.NewService(ids, logger)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// untrust any proxy while developing
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Error("set trusted proxies failed", "error", err)
		os.Exit(1)
	}

	accountHandler := web.NewAccountHandler(accountService, logger)
	accountHandler.RegisterRoutes(router)

	logger.Info("server started", "addr", ":8080")

	if err := router.Run(":8080"); err != nil {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
