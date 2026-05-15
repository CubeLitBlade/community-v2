// package main starts the API server for the community-v2 backend
package main

import (
	"fmt"
	"log"

	"github.com/CubeLitBlade/community-v2/backend/internal/bootstrap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
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
