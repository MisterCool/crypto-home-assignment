package crypto

import (
	"context"
	"fmt"
	"time"

	"deblock-home-assignment/internal/model"
	"github.com/go-resty/resty/v2"
)

type SolanaFetcher struct {
	client      *resty.Client
	rateLimiter <-chan time.Time
}

func NewSolanaFetcher(rpcURL, apiKey string) *SolanaFetcher {
	client := resty.New().
		SetBaseURL(rpcURL).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)
	if apiKey != "" {
		client.SetHeader("X-API-Key", apiKey)
	}
	return &SolanaFetcher{
		client:      client,
		rateLimiter: time.Tick(200 * time.Millisecond), // ~5 RPS
	}
}

func (s *SolanaFetcher) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	<-s.rateLimiter
	var res struct {
		Result int64 `json:"result"`
	}
	err := s.rpcCall(ctx, "getSlot", []interface{}{}, &res)
	return res.Result, err
}

func (s *SolanaFetcher) GetBlockTxs(ctx context.Context, blockNumber int64) (*model.RawBlock, error) {
	<-s.rateLimiter
	var res struct {
		Result struct {
			Transactions []struct {
				Transaction struct {
					Signatures []string `json:"signatures"`
				} `json:"transaction"`
			} `json:"transactions"`
		} `json:"result"`
	}
	if err := s.rpcCall(ctx, "getBlock", []interface{}{blockNumber}, &res); err != nil {
		return nil, err
	}

	txList := make([]model.RawTx, 0)
	for _, tx := range res.Result.Transactions {
		if len(tx.Transaction.Signatures) > 0 {
			txList = append(txList, model.RawTx{TxID: tx.Transaction.Signatures[0]})
		}
	}

	return &model.RawBlock{
		BlockNumber: blockNumber,
		TxIDs:       txList,
	}, nil
}

func (s *SolanaFetcher) GetTxDetails(ctx context.Context, txID string) (model.RawTx, error) {
	<-s.rateLimiter
	var res struct {
		Result struct {
			Transaction struct {
				Message struct {
					AccountKeys []string `json:"accountKeys"`
				} `json:"message"`
				Signatures []string `json:"signatures"`
			} `json:"transaction"`
			Meta struct {
				PostBalances []int64 `json:"postBalances"`
				PreBalances  []int64 `json:"preBalances"`
			} `json:"meta"`
		} `json:"result"`
	}
	if err := s.rpcCall(ctx, "getConfirmedTransaction", []interface{}{txID, "json"}, &res); err != nil {
		return model.RawTx{}, err
	}

	from := ""
	to := ""
	amount := 0.0
	if len(res.Result.Transaction.Message.AccountKeys) >= 2 && len(res.Result.Meta.PreBalances) >= 2 && len(res.Result.Meta.PostBalances) >= 2 {
		from = res.Result.Transaction.Message.AccountKeys[0]
		to = res.Result.Transaction.Message.AccountKeys[1]
		diff := res.Result.Meta.PreBalances[0] - res.Result.Meta.PostBalances[0]
		amount = float64(diff) / 1e9
	}

	return model.RawTx{
		TxID: txID,
		Vin: []model.Vin{{
			Address: from,
		}},
		Vout: []model.Vout{{
			Value: amount,
			ScriptPubKey: model.ScriptPubKey{
				Address: to,
			},
		}},
	}, nil
}

func (s *SolanaFetcher) rpcCall(ctx context.Context, method string, params []interface{}, out interface{}) error {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "go-client",
		"method":  method,
		"params":  params,
	}
	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(out).
		Post("")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("rpc error: %s", resp.String())
	}
	return nil
}
