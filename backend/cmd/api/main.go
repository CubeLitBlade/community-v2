package main

import (
	"fmt"
	"log"
	"os"

	"github.com/CubeLitBlade/community-v2/backend/internal/bootstrap"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	app, err := bootstrap.NewApp()
	if err != nil {
		return fmt.Errorf("bootstrap app: %w", err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
