package config

import (
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string

	KeycloakIssuer       string
	KeycloakClientID     string
	KeycloakClientSecret string

	Database PostgresConfig
	Contract ContractConfig
	Kafka    KafkaConfig
	Outbox   OutboxConfig
}

func Load() (Config, error) {
	_ = godotenv.Load()

	dbConfig, err := LoadPostgresConfig()
	if err != nil {
		return Config{}, err
	}

	contractConfig, err := LoadContractConfig()
	if err != nil {
		return Config{}, err
	}

	kafkaConfig, err := LoadKafkaConfig()
	if err != nil {
		return Config{}, err
	}

	outboxConfig, err := LoadOutboxConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPPort:             getenv("HTTP_PORT", "8080"),
		KeycloakIssuer:       getenv("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip"),
		KeycloakClientID:     getenv("KEYCLOAK_CLIENT_ID", "sharetrip-api"),
		KeycloakClientSecret: getenv("KEYCLOAK_CLIENT_SECRET", "kcTclgACcVx4ozusKmvvihUqARRE4OnI"),

		Database: dbConfig,
		Contract: contractConfig,
		Kafka:    kafkaConfig,
		Outbox:   outboxConfig,
	}, nil
}
