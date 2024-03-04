package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

func (p *postgres) AddBlock(ctx context.Context, block repository.Block) error {
	query, args, err := p.psql.Insert(blocksTable).
		Columns("block_id", "header_version_app", "header_version_block", "header_chain_id", "header_height", "header_time",
			"header_last_block_id_hash", "header_last_block_id_parts_header_total", "header_last_block_id_parts_header_hash", "header_last_commit_hash",
			"header_data_hash", "header_validators_hash", "header_next_validators_hash", "header_consensus_hash", "header_app_hash",
			"header_last_results_hash", "header_evidence_hash", "header_proposer_address",
			"commit_height", "commit_round", "commit_block_id_hash", "commit_block_id_parts_header_total", "commit_block_id_parts_header_hash").
		Values(block.BlockID, block.HeaderVersionApp, block.HeaderVersionBlock, block.HeaderChainID, block.HeaderHeight, block.HeaderTime,
			block.HeaderLastBlockIDHash, block.HeaderLastBlockIDPartsHeaderTotal, block.HeaderLastBlockIDPartsHeaderHash, block.HeaderLastCommitHash,
			block.HeaderDataHash, block.HeaderValidatorsHash, block.HeaderNextValidatorsHash, block.HeaderConsensusHash, block.HeaderAppHash,
			block.HeaderLastResultsHash, block.HeaderEvidenceHash, block.HeaderProposerAddress,
			block.CommitHeight, block.CommitRound, block.CommitBlockIDHash, block.CommitBlockIDPartsHeaderTotal, block.CommitBlockIDPartsHeaderHash).
		ToSql()
	if err != nil {
		return errors.New(err, "Build SQL for AddBlock")
	}

	_, err = p.exec.ExecContext(ctx, query, args...)
	return errors.New(err, "Exec SQL for AddBlock")
}

func (p *postgres) GetLastHeight(ctx context.Context) (int64, error) {
	query, args, err := p.psql.Select("COALESCE(MAX(header_height), 0)").From(blocksTable).ToSql()
	if err != nil {
		return 0, errors.New(err, "Build SQL for GetLastHeight")
	}

	var height int64
	err = p.exec.QueryRowContext(ctx, query, args...).Scan(&height)
	if err == sql.ErrNoRows {
		return 0, nil
	}

	return height, errors.New(err, "Exec SQL for GetLastHeight")
}

func (p *postgres) GetBlockBy(ctx context.Context, filter repository.BlockFilter) (repository.Block, error) {
	builder := p.psql.Select("block_id", "header_version_app", "header_version_block", "header_chain_id", "header_height", "header_time",
		"header_last_block_id_hash", "header_last_block_id_parts_header_total", "header_last_block_id_parts_header_hash", "header_last_commit_hash",
		"header_data_hash", "header_validators_hash", "header_next_validators_hash", "header_consensus_hash", "header_app_hash",
		"header_last_results_hash", "header_evidence_hash", "header_proposer_address",
		"commit_height", "commit_round", "commit_block_id_hash", "commit_block_id_parts_header_total", "commit_block_id_parts_header_hash").
		From(blocksTable)

	if filter.Height != 0 {
		builder = builder.Where(sq.Eq{"header_height": filter.Height})
	}
	if len(filter.BlockID) != 0 {
		builder = builder.Where(sq.Eq{"block_id": filter.BlockID})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return repository.Block{}, errors.New(err, "Build SQL for GetBlockBy")
	}

	var block repository.Block
	err = p.exec.QueryRowContext(ctx, query, args...).Scan(&block.BlockID, &block.HeaderVersionApp, &block.HeaderVersionBlock, &block.HeaderChainID, &block.HeaderHeight, &block.HeaderTime,
		&block.HeaderLastBlockIDHash, &block.HeaderLastBlockIDPartsHeaderTotal, &block.HeaderLastBlockIDPartsHeaderHash,
		&block.HeaderLastCommitHash, &block.HeaderDataHash, &block.HeaderValidatorsHash, &block.HeaderNextValidatorsHash, &block.HeaderConsensusHash, &block.HeaderAppHash,
		&block.HeaderLastResultsHash, &block.HeaderEvidenceHash, &block.HeaderProposerAddress,
		&block.CommitHeight, &block.CommitRound, &block.CommitBlockIDHash, &block.CommitBlockIDPartsHeaderTotal, &block.CommitBlockIDPartsHeaderHash)
	if err == sql.ErrNoRows {
		return block, repository.ErrNotFound
	}

	return block, errors.New(err, "Exec SQL for GetBlockBy")
}

func (p *postgres) GetLatestBlocks(ctx context.Context, cnt, offset uint64) ([]*repository.Block, error) {
	builder := p.psql.Select("block_id", "header_version_app", "header_version_block", "header_chain_id", "header_height", "header_time",
		"header_last_block_id_hash", "header_last_block_id_parts_header_total", "header_last_block_id_parts_header_hash", "header_last_commit_hash",
		"header_data_hash", "header_validators_hash", "header_next_validators_hash", "header_consensus_hash", "header_app_hash",
		"header_last_results_hash", "header_evidence_hash", "header_proposer_address",
		"commit_height", "commit_round", "commit_block_id_hash", "commit_block_id_parts_header_total", "commit_block_id_parts_header_hash").
		From(blocksTable).OrderBy("header_height DESC")

	if cnt != 0 {
		builder = builder.Limit(cnt)
	}
	builder = builder.Offset(offset)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetLatestBlocks")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetLatestBlocks")
	}
	defer rows.Close()

	var blocks []*repository.Block
	for rows.Next() {
		var block repository.Block
		if err = rows.Scan(&block.BlockID, &block.HeaderVersionApp, &block.HeaderVersionBlock, &block.HeaderChainID, &block.HeaderHeight, &block.HeaderTime,
			&block.HeaderLastBlockIDHash, &block.HeaderLastBlockIDPartsHeaderTotal, &block.HeaderLastBlockIDPartsHeaderHash,
			&block.HeaderLastCommitHash, &block.HeaderDataHash, &block.HeaderValidatorsHash, &block.HeaderNextValidatorsHash, &block.HeaderConsensusHash, &block.HeaderAppHash,
			&block.HeaderLastResultsHash, &block.HeaderEvidenceHash, &block.HeaderProposerAddress,
			&block.CommitHeight, &block.CommitRound, &block.CommitBlockIDHash, &block.CommitBlockIDPartsHeaderTotal, &block.CommitBlockIDPartsHeaderHash); err != nil {
			return nil, errors.New(err, "Scan result for GetLatestBlocks")
		}
		blocks = append(blocks, &block)
	}

	return blocks, nil
}
