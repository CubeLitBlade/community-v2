package bootstrap

import (
	"fmt"

	fga "github.com/openfga/go-sdk/client"
	"go.uber.org/fx"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/shared"
)

func provideOpenFGAClient(cfg *shared.OpenFGAConfig) (*fga.OpenFgaClient, error) {
	client, err := fga.NewSdkClient(&fga.ClientConfiguration{
		ApiUrl:               cfg.APIURL,
		StoreId:              cfg.StoreID,
		AuthorizationModelId: cfg.ModelID,
	})
	if err != nil {
		return nil, fmt.Errorf("create openfga client: %w", err)
	}

	return client, nil
}

// OpenFGAModule provides the OpenFGA client and its dependencies via fx.
func OpenFGAModule() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(provideOpenFGAClient,
				fx.As(new(authz.SDKWriter)),
				fx.As(new(authz.SDKChecker)),
				fx.As(new(authz.SDKStore)),
			)),
	)
}
