package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

func (p *postgres) AddAccountTransactions(ctx context.Context, txs ...repository.AccountTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	builder := p.psql.Insert(accountTransactionsTable).
		Columns("address", "tx_hash", "block_height", "tx_pos")

	for _, tx := range txs {
		builder = builder.Values(tx.Address, tx.TxHash, tx.BlockHeight, tx.TxPos)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.New(err, "Build SQL for AddAccountTransactions")
	}

	_, err = p.exec.ExecContext(ctx, query, args...)
	return errors.New(err, "Exec SQL for AddAccountTransactions")
}

func (p *postgres) GetTotalAccountTxs(ctx context.Context, address []byte) (uint64, error) {
	query, args, err := p.psql.Select("COUNT(*)").
		From(accountTransactionsTable).
		Where(sq.Eq{"address": address}).
		ToSql()
	if err != nil {
		return 0, errors.New(err, "Build SQL for GetTotalAccountTxs")
	}

	var total uint64
	err = p.exec.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, errors.New(err, "Exec SQL for GetTotalAccountTxs")
	}

	return total, nil
}

func (p *postgres) GetAccountTxs(ctx context.Context, address []byte, limit, offset uint64) ([][]byte, error) {
	builder := p.psql.Select("tx_hash").From(accountTransactionsTable).Where(sq.Eq{"address": address}).OrderBy("block_height DESC", "tx_pos DESC")

	if limit != 0 {
		builder = builder.Limit(limit)
	}
	builder = builder.Offset(offset)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetAccountTxs")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetAccountTxs")
	}
	defer rows.Close()

	var txHashes [][]byte
	for rows.Next() {
		var txHash []byte
		if err = rows.Scan(&txHash); err != nil {
			return nil, errors.New(err, "Scan result for GetTxsBy")
		}
		txHashes = append(txHashes, txHash)
	}

	return txHashes, nil
}

func (p *postgres) GetAccountThresholds(ctx context.Context, updateAccountCode []byte, accountID string) ([]*uint8, error) {
	query, args, err := p.psql.Select("data->>'threshold'").
		From(transactionsTable).
		Where(sq.Eq{"code": updateAccountCode}).
		Where(sq.Eq{"data->>'address'": accountID}).
		ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetAccountThresholds")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetAccountThresholds")
	}
	defer rows.Close()

	var thresholds []*uint8

	for rows.Next() {
		var threshold *uint8
		if err = rows.Scan(&threshold); err != nil {
			return nil, errors.New(err, "Scan rows for GetAccountThresholds")
		}
		thresholds = append(thresholds, threshold)
	}

	return thresholds, nil
}

func (p *postgres) GetAccountVPCodes(ctx context.Context, updateAccountCode []byte, accountID string) ([]*string, error) {
	query, args, err := p.psql.Select("data->>'vp_code_hash'").
		From(transactionsTable).
		Where(sq.Eq{"code": updateAccountCode}).
		Where(sq.Eq{"data->>'address'": accountID}).
		ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetAccountVPCodes")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetAccountVPCodes")
	}
	defer rows.Close()

	var vpCodes []*string

	for rows.Next() {
		var vpCode *string
		if err = rows.Scan(&vpCode); err != nil {
			return nil, errors.New(err, "Scan rows for GetAccountVPCodes")
		}
		vpCodes = append(vpCodes, vpCode)
	}

	return vpCodes, nil
}

func (p *postgres) GetAccountPublicKeys(ctx context.Context, updateAccountCode []byte, accountID string) ([][]string, error) {
	query, args, err := p.psql.Select("ARRAY(SELECT jsonb_array_elements_text(data->'public_keys'))").
		From(transactionsTable).
		Where(sq.Eq{"code": updateAccountCode}).
		Where(sq.Eq{"data->>'address'": accountID}).
		ToSql()
	if err != nil {
		return nil, errors.New(err, "Build SQL for GetAccountPublicKeys")
	}

	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(err, "Exec SQL for GetAccountPublicKeys")
	}
	defer rows.Close()

	var publicKeys [][]string

	for rows.Next() {
		var keys []string
		if err = rows.Scan(pq.Array(&keys)); err != nil {
			return nil, errors.New(err, "Scan rows for GetAccountPublicKeys")
		}
		publicKeys = append(publicKeys, keys)
	}

	return publicKeys, nil
}
