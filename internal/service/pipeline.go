package service

import (
	"context"
	"fmt"
	"time"

	"deblock-home-assignment/internal/config"
	data_layer "deblock-home-assignment/internal/data-layer"
	"deblock-home-assignment/internal/data-layer/crypto"
	"deblock-home-assignment/internal/model"
	"deblock-home-assignment/internal/service/kafka"
)

type ChainConfig struct {
	Name       model.Chain
	Fetcher    data_layer.BlockFetcher
	StartBlock int64
	BatchSize  int64
	Interval   time.Duration
}

func RunPipelineFromYAML(ctx context.Context, cfg *config.AppConfig, producer kafka.Producer, addressToUser map[string]string) {
	var chains []ChainConfig

	for _, c := range cfg.Chains {
		chain := model.Chain(c.Chain)

		var fetcher data_layer.BlockFetcher
		switch chain {
		case model.ChainBitcoin:
			fetcher = crypto.NewBitcoinFetcher(c.RPCUrl, c.APIKey)
		case model.ChainEthereum:
			fetcher = crypto.NewEthereumFetcher(c.RPCUrl, c.APIKey)
		case model.ChainSolana:
			fetcher = crypto.NewSolanaFetcher(c.RPCUrl, c.APIKey)
		default:
			fmt.Printf("unsupported chain: %s\n", c.Chain)
			continue
		}

		chains = append(chains, ChainConfig{
			Name:       chain,
			Fetcher:    fetcher,
			StartBlock: c.StartFrom,
			BatchSize:  c.BatchSize,
			Interval:   10 * time.Second,
		})
	}

	RunPipeline(ctx, producer, addressToUser, chains)
}

func RunPipeline(ctx context.Context, producer kafka.Producer, addressToUser map[string]string, configs []ChainConfig) {
	for _, cfg := range configs {
		coordinator := &BatchCoordinator{
			FromBlock:    cfg.StartBlock,
			BatchSize:    cfg.BatchSize,
			PollInterval: cfg.Interval,
			Fetcher:      cfg.Fetcher,
			Out:          make(chan model.BlockRange, 100),
		}

		go coordinator.Run(ctx)

		rawTxChan := make(chan []model.RawTx, 100)
		go StartChainFetcher(ctx, cfg.Name, coordinator.Out, coordinator.Fetcher, rawTxChan)

		filteredChan := StartTxFilter(ctx, cfg.Name, rawTxChan, addressToUser, coordinator.Fetcher)

		go func(chain model.Chain, ch <-chan *model.FilteredTxEvent) {
			for tx := range ch {
				if err := producer.Publish(ctx, tx); err != nil {
					fmt.Printf("[%s] failed to publish tx: %v\n", chain, err)
				}
			}
		}(cfg.Name, filteredChan)
	}
}
