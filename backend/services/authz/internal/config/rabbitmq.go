package config

// RabbitMQConfig holds the configuration for the RabbitMQ message broker connection.
type RabbitMQConfig struct {
	BrokerURL string `koanf:"broker_url"`
}
