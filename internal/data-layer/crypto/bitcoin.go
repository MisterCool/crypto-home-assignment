package crypto

import (
	"context"
	"fmt"
	"time"

	"deblock-home-assignment/internal/model"
	"github.com/go-resty/resty/v2"
)

type BitcoinFetcher struct {
	client      *resty.Client
	rateLimiter <-chan time.Time
}

func NewBitcoinFetcher(rpcURL, apiKey string) *BitcoinFetcher {
	client := resty.New().
		SetBaseURL(rpcURL).
		SetHeader("Content-Type", "application/json")
	if apiKey != "" {
		client.SetHeader("X-API-Key", apiKey)
	}
	return &BitcoinFetcher{client: client}
}

func (b *BitcoinFetcher) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	<-b.rateLimiter
	var res struct {
		Result int64 `json:"result"`
	}
	err := b.rpcCall(ctx, "getblockcount", []interface{}{}, &res)
	return res.Result, err
}

func (b *BitcoinFetcher) GetBlockTxs(ctx context.Context, blockNumber int64) (*model.RawBlock, error) {
	<-b.rateLimiter
	var hashRes struct {
		Result string `json:"result"`
	}
	if err := b.rpcCall(ctx, "getblockhash", []interface{}{blockNumber}, &hashRes); err != nil {
		return nil, err
	}

	var blockRes struct {
		Result struct {
			Tx []struct {
				Txid string       `json:"txid"`
				Vin  []model.Vin  `json:"vin"`
				Vout []model.Vout `json:"vout"`
			} `json:"tx"`
		} `json:"result"`
	}
	if err := b.rpcCall(ctx, "getblock", []interface{}{hashRes.Result, 2}, &blockRes); err != nil {
		return nil, err
	}

	txs := make([]model.RawTx, 0, len(blockRes.Result.Tx))
	for _, t := range blockRes.Result.Tx {
		txs = append(txs, model.RawTx{
			TxID: t.Txid,
			Vin:  t.Vin,
			Vout: t.Vout,
		})
	}

	return &model.RawBlock{
		BlockNumber: blockNumber,
		TxIDs:       txs,
	}, nil
}

func (b *BitcoinFetcher) GetTxDetails(ctx context.Context, txid string) (model.RawTx, error) {
	<-b.rateLimiter
	type prevout struct {
		Value               float64 `json:"value"`
		ScriptPubKeyAddress string  `json:"scriptpubkey_address"`
	}
	type vin struct {
		Prevout *prevout `json:"prevout"`
	}
	type vout struct {
		Value               float64 `json:"value"`
		ScriptPubKeyAddress string  `json:"scriptpubkey_address"`
	}
	type result struct {
		TxID string `json:"txid"`
		Vin  []vin  `json:"vin"`
		Vout []vout `json:"vout"`
	}
	var res struct {
		Result result `json:"result"`
	}

	err := b.rpcCall(ctx, "getrawtransaction", []interface{}{txid, true}, &res)
	if err != nil {
		return model.RawTx{}, err
	}

	var (
		totalVin  float64
		totalVout float64
		from      string
		to        string
	)

	for _, in := range res.Result.Vin {
		if in.Prevout != nil {
			totalVin += in.Prevout.Value
			if from == "" {
				from = in.Prevout.ScriptPubKeyAddress
			}
		}
	}

	for _, out := range res.Result.Vout {
		totalVout += out.Value
		if to == "" {
			to = out.ScriptPubKeyAddress
		}
	}

	return model.RawTx{}, nil
}

func (b *BitcoinFetcher) rpcCall(ctx context.Context, method string, params []interface{}, out interface{}) error {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "go-client",
		"method":  method,
		"params":  params,
	}
	resp, err := b.client.R().
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
