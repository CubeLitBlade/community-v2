// Package main starts the API server for the authz service.
package main

import (
	"log"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	app.Run()

	log.Println("server stopped")
}
