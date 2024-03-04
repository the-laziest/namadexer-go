package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

func (p *postgres) AddCommitSignatures(ctx context.Context, signatures ...repository.CommitSignature) error {
	if len(signatures) == 0 {
		return nil
	}

	builder := p.psql.Insert(commitSignaturesTable).
		Columns("block_id", "block_id_flag", "validator_address", "timestamp", "signature")

	for _, signature := range signatures {
		builder = builder.Values(signature.BlockID, signature.BlockIDFlag, signature.ValidatorAddress, signature.Timestamp, signature.Signature)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.New(err, "Build SQL for AddCommitSignatures")
	}

	_, err = p.exec.ExecContext(ctx, query, args...)
	return errors.New(err, "Exec SQL for AddCommitSignatures")
}

func (p *postgres) GetCommitsCount(ctx context.Context, validatorAddress []byte, start, end int64) (int64, error) {
	query, args, err := p.psql.Select("COUNT(*)").
		From(commitSignaturesTable).
		Join(blocksTable + " USING (block_id)").
		Where(sq.Eq{"validator_address": validatorAddress}).
		Where(sq.GtOrEq{"header_height": start}).
		Where(sq.LtOrEq{"header_height": end}).
		ToSql()
	if err != nil {
		return 0, errors.New(err, "Build SQL for GetCommitsCount")
	}

	var cnt int64
	err = p.exec.QueryRowContext(ctx, query, args...).Scan(&cnt)
	if err == sql.ErrNoRows {
		return 0, repository.ErrNotFound
	}

	return cnt, errors.New(err, "Exec SQL for GetCommitsCount")
}
