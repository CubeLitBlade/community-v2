package bootstrap

import (
	"fmt"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz"
	fga "github.com/openfga/go-sdk/client"
	"go.uber.org/fx"
)

// OpenFGAConfig holds the configuration required to connect to the OpenFGA server.
type OpenFGAConfig struct {
	URL     string `env:"FGA_API_URL" required:"true"`
	StoreID string `env:"FGA_STORE_ID" required:"true"`
	ModelID string `env:"FGA_MODEL_ID" required:"true"`
}

func provideOpenFGAClient(cfg OpenFGAConfig) (*fga.OpenFgaClient, error) {
	client, err := fga.NewSdkClient(&fga.ClientConfiguration{
		ApiUrl:               cfg.URL,
		StoreId:              cfg.StoreID,
		AuthorizationModelId: cfg.ModelID,
	})
	if err != nil {
		return nil, fmt.Errorf("create openfga client: %w", err)
	}

	return client, nil
}

// FGAModule provides the OpenFGA client and its dependencies via fx.
func FGAModule() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(provideOpenFGAClient,
				fx.As(new(authz.SDKWriter)),
				fx.As(new(authz.SDKChecker)),
				fx.As(new(authz.SDKStore)),
			)),
	)
}
