package service

import (
	"context"
	"fmt"

	data_layer "deblock-home-assignment/internal/data-layer"
	"deblock-home-assignment/internal/model"
)

func StartTxFilter(
	ctx context.Context,
	chain model.Chain,
	in <-chan []model.RawTx,
	addressToUserID map[string]string,
	fetcher data_layer.BlockFetcher,
) <-chan *model.FilteredTxEvent {
	out := make(chan *model.FilteredTxEvent, 100)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case txBatch, ok := <-in:
				if !ok {
					return
				}
				for _, tx := range txBatch {
					for _, vout := range tx.Vout {
						addr := vout.ScriptPubKey.Address
						userID, ok := addressToUserID[addr]

						if !ok || addr == "" {
							continue
						}

						fee, err := CalculateFee(ctx, tx, fetcher)
						if err != nil {
							fmt.Println(err)
							fee = 0
						}

						source := ""
						if len(tx.Vin) > 0 {
							source = tx.Vin[0].TxID
						}

						out <- &model.FilteredTxEvent{
							UserID: userID,
							Chain:  chain,
							TxHash: tx.TxID,
							From:   source,
							To:     addr,
							Amount: vout.Value,
							Fee:    fee,
						}
					}
				}
			}
		}
	}()

	return out
}

func CalculateFee(ctx context.Context, tx model.RawTx, fetcher data_layer.BlockFetcher) (float64, error) {
	var vinSum float64

	for _, vin := range tx.Vin {
		if vin.TxID == "" {
			continue
		}

		prevTx, err := fetcher.GetTxDetails(ctx, vin.TxID)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch prev tx %s: %w", vin.TxID, err)
		}

		for _, v := range prevTx.Vout {
			if v.N == vin.Vout {
				vinSum += v.Value
				break
			}
		}
	}

	var voutSum float64
	for _, v := range tx.Vout {
		voutSum += v.Value
	}

	fee := vinSum - voutSum
	if fee < 0 {
		fee = 0
	}

	return fee, nil
}
