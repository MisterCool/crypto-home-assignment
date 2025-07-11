package service

import (
	"context"
	"log"
	"time"

	data_layer "deblock-home-assignment/internal/data-layer"
	"deblock-home-assignment/internal/model"
)

type BatchCoordinator struct {
	FromBlock    int64
	BatchSize    int64
	PollInterval time.Duration
	Fetcher      data_layer.BlockFetcher
	Out          chan model.BlockRange
}

func (c *BatchCoordinator) Run(ctx context.Context) {
	ticker := time.NewTicker(c.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest, err := c.Fetcher.GetLatestBlockHeight(ctx)
			if err != nil {
				log.Printf("[Coordinator] error fetching latest block: %v", err)
				continue
			}

			if latest < c.FromBlock {
				continue
			}

			to := c.FromBlock + c.BatchSize - 1
			if to > latest {
				to = latest
			}

			br := model.BlockRange{From: c.FromBlock, To: to}
			log.Printf("[Coordinator] sending block range: %d → %d", br.From, br.To)
			c.Out <- br
			c.FromBlock = to + 1
		}
	}
}
