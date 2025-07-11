package kafka

import (
	"context"
	"fmt"

	"deblock-home-assignment/internal/model"
)

type MockProducer struct{}

func (m *MockProducer) Publish(ctx context.Context, tx *model.FilteredTxEvent) error {
	fmt.Printf("[MOCK-KAFKA] Chain: %s | User: %s | From: %s → To: %s | Amount: %.8f | Fee: %.8f\n",
		tx.Chain, tx.UserID, tx.From, tx.To, tx.Amount, tx.Fee)
	return nil
}
