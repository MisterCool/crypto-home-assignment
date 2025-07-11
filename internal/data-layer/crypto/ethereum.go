package crypto

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"deblock-home-assignment/internal/model"
	"github.com/go-resty/resty/v2"
)

type EthereumFetcher struct {
	client      *resty.Client
	rateLimiter <-chan time.Time
}

func NewEthereumFetcher(rpcURL, apiKey string) *EthereumFetcher {
	client := resty.New().
		SetBaseURL(rpcURL).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)
	if apiKey != "" {
		client.SetHeader("X-API-Key", apiKey)
	}
	return &EthereumFetcher{client: client}
}

func (e *EthereumFetcher) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	<-e.rateLimiter
	var res struct {
		Result string `json:"result"`
	}
	if err := e.rpcCall(ctx, "eth_blockNumber", []interface{}{}, &res); err != nil {
		return 0, err
	}
	var height int64
	_, err := fmt.Sscanf(res.Result, "0x%x", &height)
	return height, err
}

func (e *EthereumFetcher) GetBlockTxs(ctx context.Context, blockNumber int64) (*model.RawBlock, error) {
	<-e.rateLimiter
	hexNum := fmt.Sprintf("0x%x", blockNumber)

	var res struct {
		Result struct {
			Transactions []struct {
				Hash  string `json:"hash"`
				From  string `json:"from"`
				To    string `json:"to"`
				Value string `json:"value"`
			} `json:"transactions"`
		} `json:"result"`
	}

	if err := e.rpcCall(ctx, "eth_getBlockByNumber", []interface{}{hexNum, true}, &res); err != nil {
		return nil, err
	}

	rawTxs := make([]model.RawTx, 0, len(res.Result.Transactions))
	for _, tx := range res.Result.Transactions {
		valWei := new(big.Int)
		if _, ok := valWei.SetString(tx.Value[2:], 16); !ok {
			return nil, fmt.Errorf("failed to parse value hex for tx %s", tx.Hash)
		}

		valEth := new(big.Float).Quo(new(big.Float).SetInt(valWei), big.NewFloat(1e18))
		valueFloat, _ := valEth.Float64()

		rawTxs = append(rawTxs, model.RawTx{
			TxID: tx.Hash,
			Vin: []model.Vin{{
				Address: tx.From,
			}},
			Vout: []model.Vout{{
				Value: valueFloat,
				ScriptPubKey: model.ScriptPubKey{
					Address: tx.To,
				},
			}},
		})
	}

	return &model.RawBlock{
		BlockNumber: blockNumber,
		TxIDs:       rawTxs,
	}, nil
}

func (e *EthereumFetcher) GetTxDetails(ctx context.Context, txID string) (model.RawTx, error) {
	<-e.rateLimiter
	var res struct {
		Result struct {
			Hash  string `json:"hash"`
			From  string `json:"from"`
			To    string `json:"to"`
			Value string `json:"value"` // hex string
		} `json:"result"`
	}
	if err := e.rpcCall(ctx, "eth_getTransactionByHash", []interface{}{txID}, &res); err != nil {
		return model.RawTx{}, err
	}

	valueWei, err := strconv.ParseInt(res.Result.Value[2:], 16, 64) // strip `0x`
	if err != nil {
		return model.RawTx{}, err
	}

	return model.RawTx{
		TxID: res.Result.Hash,
		Vin: []model.Vin{{
			Address: res.Result.From,
		}},
		Vout: []model.Vout{{
			Value: float64(valueWei) / 1e18,
			ScriptPubKey: model.ScriptPubKey{
				Address: res.Result.To,
			},
		}},
	}, nil
}

func (e *EthereumFetcher) rpcCall(ctx context.Context, method string, params []interface{}, out interface{}) error {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "go-client",
		"method":  method,
		"params":  params,
	}
	resp, err := e.client.R().
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
