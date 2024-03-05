package service

import (
	"context"

	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
	"github.com/the-laziest/namadexer-go/internal/repository"
)

func repoBlockToInfo(block *repository.Block) BlockInfo {
	header := tmtypes.Header{
		Version:            version.Consensus{Block: block.HeaderVersionBlock, App: block.HeaderVersionApp},
		ChainID:            block.HeaderChainID,
		Height:             block.HeaderHeight,
		Time:               block.HeaderTime,
		LastBlockID:        tmtypes.BlockID{Hash: block.HeaderLastBlockIDHash, PartSetHeader: tmtypes.PartSetHeader{Total: block.HeaderLastBlockIDPartsHeaderTotal, Hash: block.HeaderLastBlockIDPartsHeaderHash}},
		LastCommitHash:     block.HeaderLastCommitHash,
		DataHash:           block.HeaderDataHash,
		ValidatorsHash:     block.HeaderValidatorsHash,
		NextValidatorsHash: block.HeaderNextValidatorsHash,
		ConsensusHash:      block.HeaderConsensusHash,
		AppHash:            block.HeaderAppHash,
		LastResultsHash:    block.HeaderLastResultsHash,
		EvidenceHash:       block.HeaderEvidenceHash,
		ProposerAddress:    block.HeaderProposerAddress,
	}
	lastCommit := LastCommitInfo{
		Height:  block.CommitHeight,
		Round:   block.CommitRound,
		BlockID: tmtypes.BlockID{Hash: block.CommitBlockIDHash, PartSetHeader: tmtypes.PartSetHeader{Total: block.CommitBlockIDPartsHeaderTotal, Hash: block.CommitBlockIDPartsHeaderHash}},
	}
	return BlockInfo{
		BlockID:    block.BlockID,
		Header:     header,
		LastCommit: &lastCommit,
	}
}

func (s *service) GetBlockByHeight(ctx context.Context, height int64) (BlockInfo, error) {
	block, err := s.repo.GetBlockBy(ctx, repository.BlockFilter{Height: height})
	if err == repository.ErrNotFound {
		return BlockInfo{}, ErrNotFound
	}
	if err != nil {
		return BlockInfo{}, err
	}

	blockInfo := repoBlockToInfo(&block)

	txs, err := s.repo.GetTxsBy(ctx, repository.TxFilter{BlockID: block.BlockID})
	if err != nil {
		return BlockInfo{}, err
	}

	blockInfo.TxHashes = repoTxsToShort(txs)

	return blockInfo, nil
}

func (s *service) GetBlockByHash(ctx context.Context, hash string) (BlockInfo, error) {
	blockID, err := hexToBytes(hash)
	if err != nil {
		return BlockInfo{}, err
	}

	block, err := s.repo.GetBlockBy(ctx, repository.BlockFilter{BlockID: blockID})
	if err == repository.ErrNotFound {
		return BlockInfo{}, ErrNotFound
	}
	if err != nil {
		return BlockInfo{}, err
	}

	blockInfo := repoBlockToInfo(&block)

	txs, err := s.repo.GetTxsBy(ctx, repository.TxFilter{BlockID: block.BlockID})
	if err != nil {
		return BlockInfo{}, err
	}

	blockInfo.TxHashes = make([]TxShort, 0, len(txs))
	for _, tx := range txs {
		blockInfo.TxHashes = append(blockInfo.TxHashes, TxShort{TxType: tx.TxType, HashID: tx.Hash})
	}

	return blockInfo, nil
}

func (s *service) GetLatestBlocks(ctx context.Context, limit, offset int64) ([]BlockInfo, error) {
	if limit <= 0 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	blocks, err := s.repo.GetLatestBlocks(ctx, uint64(limit), uint64(offset))
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, ErrNotFound
	}

	infos := make([]BlockInfo, 0, len(blocks))
	for _, block := range blocks {
		info := repoBlockToInfo(block)

		txs, err := s.repo.GetTxsBy(ctx, repository.TxFilter{BlockID: block.BlockID})
		if err != nil {
			return nil, err
		}

		info.TxHashes = make([]TxShort, 0, len(txs))
		for _, tx := range txs {
			info.TxHashes = append(info.TxHashes, TxShort{TxType: tx.TxType, HashID: tx.Hash})
		}

		infos = append(infos, info)
	}

	return infos, nil
}
