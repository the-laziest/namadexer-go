package service

import (
	"context"
	"encoding/json"

	"github.com/the-laziest/namadexer-go/pkg/errors"
)

type Service interface {
	GetBlockByHeight(ctx context.Context, height int64) (BlockInfo, error)
	GetBlockByHash(ctx context.Context, hash string) (BlockInfo, error)
	GetLatestBlocks(ctx context.Context, limit, offset int64) ([]BlockInfo, error)

	GetTxsByHashes(ctx context.Context, hash ...string) ([]TxInfo, error)
	GetTxsByMemo(ctx context.Context, memo string, limit, offset int64) ([]TxShort, error)
	GetTxsByAccount(ctx context.Context, addressHex string, limit, offset int64) ([]Hash, error)

	GetTotalTxsByMemo(ctx context.Context, memo string) (Total, error)
	GetTotalTxsByAccount(ctx context.Context, addressHex string) (Total, error)

	GetShielded(ctx context.Context) (ShieldedAssets, error)
	GetValidatorsUptime(ctx context.Context, validator string, start, end int64) (Uptime, error)
	GetVoteProposalData(ctx context.Context, proposalID int64) ([]json.RawMessage, error)
	GetAccountUpdates(ctx context.Context, accountID string) (*AccountUpdates, error)
}

var (
	ErrBadRequest = errors.Create("bad request")
	ErrNotFound   = errors.Create("resource not found")
)
