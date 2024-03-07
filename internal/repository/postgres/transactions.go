package postgres

import (
	"context"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

func (p *postgres) AddTransactions(ctx context.Context, txs ...repository.Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	builder := p.psql.Insert(transactionsTable).
		Columns("hash", "block_id", "tx_type", "wrapper_id", "memo", "fee_amount_per_gas_unit", "fee_token", "gas_limit_multiplier", "code", "data", "return_code", "pos_in_block")

	for _, tx := range txs {
		builder = builder.Values(tx.Hash, tx.BlockID, tx.TxType, tx.WrapperID, tx.Memo, tx.FeeAmountPerGasUnit, tx.FeeToken, tx.GasLimitMultiplier, tx.Code, tx.Data, tx.ReturnCode, tx.PosInBlock)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.New(err, "Build SQL for AddTransactions")
	}

	_, err = p.exec.ExecContext(ctx, query, args...)
	return errors.New(err, "Exec SQL for AddTransactions")
}

func (p *postgres) GetTotalTxsBy(ctx context.Context, filter repository.TxFilter) (uint64, error) {
	builder := p.psql.Select("COUNT(*)").From(transactionsTable)

	if len(filter.Hashes) != 0 {
		builder = builder.Where(sq.Eq{"hash": filter.Hashes})
	}
	if len(filter.BlockID) != 0 {
		builder = builder.Where(sq.Eq{"block_id": filter.BlockID})
	}
	if filter.Memo != "" {
		builder = builder.Where(sq.Eq{"memo": filter.Memo})
	}
	if filter.Limit != 0 {
		builder = builder.Limit(filter.Limit)
	}
	builder = builder.Offset(filter.Offset)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.New(err, "Build SQL for GetTotalTxsBy")
	}

	var total uint64
	err = p.exec.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, errors.New(err, "Exec SQL for GetTotalTxsBy")
	}

	return total, nil
}

func (p *postgres) GetTxsBy(ctx context.Context, filter repository.TxFilter) ([]repository.Transaction, error) {
	builder := p.psql.Select("hash", "block_id", "tx_type", "wrapper_id", "memo", "fee_amount_per_gas_unit", "fee_token", "gas_limit_multiplier", "code", "data", "return_code", "pos_in_block").
		From(transactionsTable).
		Join(blocksTable + " USING (block_id)")

	if len(filter.Hashes) != 0 {
		builder = builder.Where(sq.Eq{"hash": filter.Hashes})
	}
	if len(filter.BlockID) != 0 {
		builder = builder.Where(sq.Eq{"block_id": filter.BlockID})
	}
	if filter.Memo != "" {
		builder = builder.Where(sq.Eq{"memo": filter.Memo})
	}
	if filter.TxType != "" {
		builder = builder.Where(sq.Eq{"tx_type": filter.TxType})
	}
	if filter.Limit != 0 {
		builder = builder.Limit(filter.Limit)
	}
	builder = builder.Offset(filter.Offset)

	builder = builder.OrderBy("header_height DESC", "pos_in_block DESC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetTxsBy")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetTxsBy: "+query)
	}
	defer rows.Close()

	var txs []repository.Transaction
	for rows.Next() {
		var tx repository.Transaction
		if err = rows.Scan(&tx.Hash, &tx.BlockID, &tx.TxType, &tx.WrapperID, &tx.Memo,
			&tx.FeeAmountPerGasUnit, &tx.FeeToken, &tx.GasLimitMultiplier, &tx.Code, &tx.Data, &tx.ReturnCode, &tx.PosInBlock); err != nil {
			return nil, errors.New(err, "Scan result for GetTxsBy")
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

func (p *postgres) GetTxsBySourceOrTarget(ctx context.Context, address string) ([]repository.Transaction, error) {
	query, args, err := p.psql.Select("hash", "block_id", "tx_type", "wrapper_id", "memo", "fee_amount_per_gas_unit", "fee_token", "gas_limit_multiplier", "code", "data", "return_code", "pos_in_block").
		From(transactionsTable).
		Where(sq.Eq{"tx_type": "Decrypted"}).
		Where(sq.Or{sq.Eq{"data ->> 'source'": address}, sq.Eq{"data ->> 'target'": address}}).
		ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetTxsBySourceOrTarget")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetTxsBySourceOrTarget")
	}
	defer rows.Close()

	var txs []repository.Transaction
	for rows.Next() {
		var tx repository.Transaction
		if err = rows.Scan(&tx.Hash, &tx.BlockID, &tx.TxType, &tx.WrapperID, &tx.Memo,
			&tx.FeeAmountPerGasUnit, &tx.FeeToken, &tx.GasLimitMultiplier, &tx.Code, &tx.Data, &tx.ReturnCode, &tx.PosInBlock); err != nil {
			return nil, errors.New(err, "Scan result for GetTxsBySourceOrTarget")
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

func (p *postgres) GetVoteProposalDatas(ctx context.Context, voteCode []byte, proposalID int64) ([]json.RawMessage, error) {
	query, args, err := p.psql.Select("data").
		From(transactionsTable).
		Join(blocksTable+" USING (block_id)").
		Where(sq.Eq{"code": voteCode}).
		Where(sq.Eq{"(data->>'id')::int": proposalID}).
		OrderBy("header_height DESC", "pos_in_block DESC").
		ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetVoteProposalDatas")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetVoteProposalDatas")
	}
	defer rows.Close()

	var datas []json.RawMessage
	for rows.Next() {
		var data json.RawMessage
		if err = rows.Scan(&data); err != nil {
			return nil, errors.New(err, "Scan result for GetVoteProposalDatas")
		}
		datas = append(datas, data)
	}

	return datas, nil
}
