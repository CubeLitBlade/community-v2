package main

import (
	"fmt"
	"os"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	app.Run()

	fmt.Fprintln(os.Stderr, "server stopped")
}
