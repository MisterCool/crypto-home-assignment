package service

import (
	"context"
	"log"

	data_layer "deblock-home-assignment/internal/data-layer"
	"deblock-home-assignment/internal/model"
)

func StartChainFetcher(
	ctx context.Context,
	chain model.Chain,
	in <-chan model.BlockRange,
	fetcher data_layer.BlockFetcher,
	txOut chan<- []model.RawTx,
) {
	workerCount := 2

	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for br := range in {
				for blockNum := br.From; blockNum <= br.To; blockNum++ {

					block, err := fetcher.GetBlockTxs(ctx, blockNum)
					if err != nil {
						log.Printf("[%s-%d] failed to fetch block %d: %v", chain, id, blockNum, err)
						continue
					}

					if len(block.TxIDs) > 0 {
						txOut <- block.TxIDs
						log.Printf("[%s-%d] block %d: %d txs sent", chain, id, blockNum, len(block.TxIDs))
					}
				}
			}
		}(i)
	}
}
