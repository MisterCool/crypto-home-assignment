package model

type (
	RawBlock struct {
		BlockNumber int64
		TxIDs       []RawTx
	}

	BlockRange struct {
		From int64
		To   int64
	}
)
