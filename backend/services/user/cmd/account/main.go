// package main starts the API server for the community-v2 backend
package main

import (
	"log"

	"github.com/CubeLitBlade/community-v2/backend/services/user/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	app.Run()

	log.Println("server stopped")
}
