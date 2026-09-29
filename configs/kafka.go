package config

import (
	"fmt"
	"os"
)

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func LoadKafkaConfig() (KafkaConfig, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return KafkaConfig{}, fmt.Errorf("KAFKA_BROKERS is required")
	}

	topic := os.Getenv("TRIP_EVENTS_TOPIC")
	if topic == "" {
		return KafkaConfig{}, fmt.Errorf("TRIP_EVENTS_TOPIC is required")
	}

	return KafkaConfig{
		Brokers: []string{brokers},
		Topic:   topic,
	}, nil
}
