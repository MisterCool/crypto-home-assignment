package model

type (
	RawTx struct {
		TxID string
		Vin  []Vin
		Vout []Vout
	}

	FilteredTxEvent struct {
		Chain  Chain
		TxHash string
		From   string
		To     string
		Amount float64
		Fee    float64
		UserID string
	}

	Vin struct {
		TxID    string `json:"txid"`
		Vout    int    `json:"vout"`
		Address string
	}

	Vout struct {
		Value        float64      `json:"value"`
		N            int          `json:"n"`
		ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
	}

	ScriptPubKey struct {
		Asm     string `json:"asm"`
		Hex     string `json:"hex"`
		ReqSigs int    `json:"reqSigs"`
		Type    string `json:"type"`
		Address string `json:"address"`
	}
)
