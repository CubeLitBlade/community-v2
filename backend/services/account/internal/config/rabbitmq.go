package config

type RabbitMQConfig struct {
	BrokerURL string `koanf:"broker_url"`
}
