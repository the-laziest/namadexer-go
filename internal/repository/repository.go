package repository

import (
	"context"
	"encoding/json"

	"github.com/the-laziest/namadexer-go/pkg/errors"
)

type Repository interface {
	CreateTables(ctx context.Context) error

	AddBlock(ctx context.Context, block Block) error
	GetBlockBy(ctx context.Context, filter BlockFilter) (Block, error)
	GetLatestBlocks(ctx context.Context, cnt, offset uint64) ([]*Block, error)

	AddTransactions(ctx context.Context, txs ...Transaction) error
	GetTotalTxsBy(ctx context.Context, filter TxFilter) (uint64, error)
	GetTxsBy(ctx context.Context, filter TxFilter) ([]Transaction, error)
	GetTxsBySourceOrTarget(ctx context.Context, address string) ([]Transaction, error)
	GetVoteProposalDatas(ctx context.Context, voteCode []byte, proposalID int64) ([]json.RawMessage, error)

	AddAccountTransactions(ctx context.Context, txs ...AccountTransaction) error
	GetTotalAccountTxs(ctx context.Context, address []byte) (uint64, error)
	GetAccountTxs(ctx context.Context, address []byte, limit, offset uint64) ([][]byte, error)

	GetAccountThresholds(ctx context.Context, updateAccountCode []byte, accountID string) ([]uint8, error)
	GetAccountVPCodes(ctx context.Context, updateAccountCode []byte, accountID string) ([]string, error)
	GetAccountPublicKeys(ctx context.Context, updateAccountCode []byte, accountID string) ([][]string, error)

	AddCommitSignatures(ctx context.Context, signatures ...CommitSignature) error
	GetCommitsCount(ctx context.Context, validatorAddress []byte, start, end int64) (int64, error)

	AddEvidences(ctx context.Context, evidences ...Evidence) error

	GetLastHeight(ctx context.Context) (int64, error)

	HasIndexes(ctx context.Context) (bool, error)
	CreateIndexes(ctx context.Context) error

	RunInTransaction(ctx context.Context, txFunc InTransaction) error

	Close() error
}

type InTransaction func(ctx context.Context, tx Repository) error

var ErrNotFound = errors.Create("no rows were found")
