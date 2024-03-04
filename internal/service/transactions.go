package service

import (
	"context"
	"encoding/json"

	"github.com/the-laziest/namadexer-go/internal/repository"
)

func repoTxToInfo(tx repository.Transaction) TxInfo {
	info := TxInfo{
		Hash:    tx.Hash,
		BlockID: tx.BlockID,
		TxType:  tx.TxType,
	}
	if len(tx.WrapperID) != 0 {
		wrapperID := Hash(tx.WrapperID)
		info.WrapperID = &wrapperID
	}
	if len(tx.FeeAmountPerGasUnit) != 0 {
		info.FeeAmountPerGasUnit = &tx.FeeAmountPerGasUnit
	}
	if len(tx.FeeToken) != 0 {
		info.FeeToken = &tx.FeeToken
	}
	if tx.GasLimitMultiplier != nil {
		info.GasLimitMultiplier = tx.GasLimitMultiplier
	}
	if len(tx.Code) != 0 {
		code := Hash(tx.Code)
		info.Code = &code
	}
	if len(tx.Data) != 0 {
		data := json.RawMessage(tx.Data)
		info.Data = &data
	}
	if tx.ReturnCode != nil {
		info.ReturnCode = tx.ReturnCode
	}
	return info
}

func repoTxsToShort(txs []repository.Transaction) []TxShort {
	txsShort := make([]TxShort, 0, len(txs))
	for _, tx := range txs {
		txsShort = append(txsShort, TxShort{TxType: tx.TxType, HashID: tx.Hash})
	}
	return txsShort
}

func (s *service) GetTxsByHashes(ctx context.Context, hashes ...string) ([]TxInfo, error) {
	if len(hashes) == 0 {
		return nil, ErrNotFound
	}

	hashIDs := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		hashID, err := hexToBytes(hash)
		if err != nil {
			return nil, err
		}
		hashIDs = append(hashIDs, hashID)
	}

	txs, err := s.repo.GetTxsBy(ctx, repository.TxFilter{Hashes: hashIDs})
	if err != nil {
		return nil, err
	}
	if len(txs) == 0 {
		return nil, ErrNotFound
	}

	txInfos := make([]TxInfo, 0, len(txs))
	for _, tx := range txs {
		txInfos = append(txInfos, repoTxToInfo(tx))
	}

	return txInfos, nil
}

func (s *service) GetTotalTxsByMemo(ctx context.Context, memo string) (Total, error) {
	total, err := s.repo.GetTotalTxsBy(ctx, repository.TxFilter{Memo: memo})
	return Total{total}, err
}

func (s *service) GetTxsByMemo(ctx context.Context, memo string, rLimit, rOffset int64) ([]TxShort, error) {
	limit, offset := prepareLimitAndOffset(rLimit, rOffset)

	txs, err := s.repo.GetTxsBy(ctx, repository.TxFilter{Memo: memo, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	return repoTxsToShort(txs), nil
}

func (s *service) GetTotalTxsByAccount(ctx context.Context, address string) (Total, error) {
	total, err := s.repo.GetTotalAccountTxs(ctx, []byte(address))
	return Total{total}, err
}

func (s *service) GetTxsByAccount(ctx context.Context, address string, rLimit, rOffset int64) ([]Hash, error) {
	limit, offset := prepareLimitAndOffset(rLimit, rOffset)

	rawHashes, err := s.repo.GetAccountTxs(ctx, []byte(address), limit, offset)
	if err != nil {
		return nil, err
	}

	hashes := make([]Hash, 0, len(rawHashes))
	for _, rawHash := range rawHashes {
		hashes = append(hashes, rawHash)
	}

	return hashes, nil
}

func (s *service) GetShielded(ctx context.Context) ([]TxInfo, error) {
	txs, err := s.repo.GetTxsBySourceOrTarget(ctx, MASP_ADDR)
	if err != nil {
		return nil, err
	}

	txInfos := make([]TxInfo, 0, len(txs))
	for _, tx := range txs {
		txInfos = append(txInfos, repoTxToInfo(tx))
	}

	return txInfos, nil
}

func prepareLimitAndOffset(limit, offset int64) (uint64, uint64) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uint64(limit), uint64(offset)
}
