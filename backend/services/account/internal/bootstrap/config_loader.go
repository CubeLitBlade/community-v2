package bootstrap

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/cubelitblade/community-v2/backend/services/account/internal/config"
)

func LoadConfig() (config.Config, error) {
	k := koanf.New(".")

	if tomlPath := os.Getenv("ACCOUNT_CONFIG_TOML_PATH"); tomlPath != "" {
		if err := k.Load(file.Provider(tomlPath), toml.Parser()); err != nil {
			return config.Config{}, fmt.Errorf("load config from toml: %w", err)
		}
	}

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "ACCOUNT_",
		TransformFunc: func(key, val string) (string, any) {
			key = strings.ReplaceAll(strings.ToLower(key), "_", ".")
			if strings.Contains(val, " ") {
				return key, strings.Split(val, " ")
			}

			return key, val
		},
	}), nil); err != nil {
		return config.Config{}, fmt.Errorf("load config from environment variables: %w", err)
	}

	var cfg config.Config
	if err := k.Unmarshal("account", &cfg); err != nil {
		return config.Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return config.Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}
