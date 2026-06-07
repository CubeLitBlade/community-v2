// Package main starts the API server for the account service.
package main

import (
	"log"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	app.Run()

	log.Println("server stopped")
}
