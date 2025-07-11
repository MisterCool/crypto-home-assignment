package kafka

import (
	"context"

	"deblock-home-assignment/internal/model"
)

type Producer interface {
	Publish(ctx context.Context, tx *model.FilteredTxEvent) error
}
