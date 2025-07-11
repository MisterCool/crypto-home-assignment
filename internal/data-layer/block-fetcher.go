package data_layer

import (
	"context"

	"deblock-home-assignment/internal/model"
)

type BlockFetcher interface {
	GetLatestBlockHeight(ctx context.Context) (int64, error)
	GetBlockTxs(ctx context.Context, blockNumber int64) (*model.RawBlock, error)
	GetTxDetails(ctx context.Context, txID string) (model.RawTx, error)
}
