package service

import (
	"encoding/hex"
	"encoding/json"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"
)

type Hash []byte

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

type LastCommitInfo struct {
	Height  int64           `json:"height"`
	Round   int32           `json:"round"`
	BlockID tmtypes.BlockID `json:"block_id"`
}

type TxShort struct {
	TxType string `json:"tx_type"`
	HashID Hash   `json:"hash_id"`
}

type BlockInfo struct {
	BlockID    Hash            `json:"block_id"`
	Header     tmtypes.Header  `json:"header"`
	LastCommit *LastCommitInfo `json:"last_commit,omitempty"`
	TxHashes   []TxShort       `json:"tx_hashes"`
}

type BlockShort struct {
	Height int64     `json:"height"`
	Time   time.Time `json:"time"`
}

type TxInfo struct {
	Hash                Hash             `json:"hash"`
	BlockID             Hash             `json:"block_id"`
	TxType              string           `json:"tx_type"`
	WrapperID           *Hash            `json:"wrapper_id,omitempty"`
	Memo                *string          `json:"memo,omitempty"`
	FeeAmountPerGasUnit *string          `json:"fee_amount_per_gas_unit,omitempty"`
	FeeToken            *string          `json:"fee_token,omitempty"`
	GasLimitMultiplier  *uint64          `json:"gas_limit_multiplier,omitempty"`
	Code                *Hash            `json:"code,omitempty"`
	Data                *json.RawMessage `json:"data,omitempty"`
	ReturnCode          *int64           `json:"return_code,omitempty"`
	BlockInfo           *BlockShort      `json:"block_info,omitempty"`
}

type Uptime struct {
	Uptime float64 `json:"uptime"`
}

type AccountUpdates struct {
	AccountID  string     `json:"account_id"`
	CodeHashes []*string  `json:"code_hashes"`
	Thresholds []*uint8   `json:"thresholds"`
	PublicKeys [][]string `json:"public_keys"`
}

type Total struct {
	Total uint64 `json:"total"`
}

type ShieldedAssets struct {
	ShieldedAssets map[string]float64 `json:"shielded_assets"`
}
