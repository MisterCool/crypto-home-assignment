package main

import (
	"context"
	"log"

	"deblock-home-assignment/internal/config"
	"deblock-home-assignment/internal/service"
	"deblock-home-assignment/internal/service/kafka"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig("cmd/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	producer := &kafka.MockProducer{}
	addressToUser := map[string]string{
		"bc1qleuzfxhc8d6qlews3dc0fu5tapmn7l6jn2s6zz": "userID1",
	}

	service.RunPipelineFromYAML(ctx, cfg, producer, addressToUser)

	select {}
}
