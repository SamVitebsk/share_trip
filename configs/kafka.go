package config

import (
	"os"
	"strings"
)

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func LoadKafkaConfig() KafkaConfig {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:29092"
	}

	topic := os.Getenv("KAFKA_PUBLISH_TRIP_TOPIC")
	if topic == "" {
		topic = "trip.events"
	}

	return KafkaConfig{
		Brokers: strings.Split(brokers, ","),
		Topic:   topic,
	}
}
