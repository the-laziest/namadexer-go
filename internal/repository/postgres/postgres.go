package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
	"github.com/the-laziest/namadexer-go/pkg/logger"
)

type postgres struct {
	config repository.Config
	db     *sql.DB
	exec   executor
	psql   sq.StatementBuilderType
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var (
	blocksTable              = "blocks"
	evidencesTable           = "evidences"
	commitSignaturesTable    = "commit_signatures"
	transactionsTable        = "transactions"
	accountTransactionsTable = "account_transactions"
)

func NewRepository(ctx context.Context, config repository.Config) (*postgres, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=%d",
		config.User, config.Password, config.Host, config.Port, config.DbName, config.ConnectionTimeout))
	if err != nil {
		return nil, errors.New(err, "Open sql connection")
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, errors.New(err, "Ping db")
	}

	config.Schema = "\"" + config.Schema + "\""
	blocksTable = config.Schema + "." + blocksTable
	evidencesTable = config.Schema + "." + evidencesTable
	commitSignaturesTable = config.Schema + "." + commitSignaturesTable
	transactionsTable = config.Schema + "." + transactionsTable
	accountTransactionsTable = config.Schema + "." + accountTransactionsTable

	return &postgres{
		config: config,
		db:     db,
		exec:   db,
		psql:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (p *postgres) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

func (p *postgres) CreateTables(ctx context.Context) error {
	createSchemaQuery := "CREATE SCHEMA IF NOT EXISTS " + p.config.Schema
	_, err := p.exec.ExecContext(ctx, createSchemaQuery)
	if err != nil {
		return errors.New(err, "Create schema")
	}

	_, err = p.exec.ExecContext(ctx, createBlocksTableQuery())
	if err != nil {
		return errors.New(err, "Create blocks table")
	}

	_, err = p.exec.ExecContext(ctx, createTransactionsTableQuery())
	if err != nil {
		return errors.New(err, "Create transactions table")
	}

	_, err = p.exec.ExecContext(ctx, createEvidencesTableQuery())
	if err != nil {
		return errors.New(err, "Create evidences table")
	}

	_, err = p.exec.ExecContext(ctx, createCommitSignaturesTableQuery())
	if err != nil {
		return errors.New(err, "Create commit signatures table")
	}

	_, err = p.exec.ExecContext(ctx, createAccountTransactionsTableQuery())
	return errors.New(err, "Create account transactions table")
}

func (p *postgres) HasIndexes(ctx context.Context) (bool, error) {
	query, args, err := p.psql.Select("indexname", "indexdef").From("pg_indexes").Where(sq.Eq{"tablename": blocksTable}).ToSql()
	if err != nil {
		return false, errors.New(err, "Has indexes build SQL")
	}
	rows, err := p.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return false, errors.New(err, "Has indexes exec SQL")
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (p *postgres) CreateIndexes(ctx context.Context) error {
	if !p.config.CreateIndex {
		return nil
	}

	blockPK := "ALTER TABLE " + blocksTable + " ADD CONSTRAINT pk_blocks_block_id PRIMARY KEY (block_id);"
	blockHeightIndex := "CREATE UNIQUE INDEX IF NOT EXISTS blocks_header_height_unique ON " + blocksTable + " (header_height);"
	txFK := "ALTER TABLE " + transactionsTable + " ADD CONSTRAINT fk_transactions_block_id FOREIGN KEY (block_id) REFERENCES " + blocksTable + " (block_id);"
	txBlockIDIndex := "CREATE INDEX IF NOT EXISTS transactions_block_id_idx ON " + transactionsTable + " USING hash(block_id);"
	txHashIndex := "CREATE INDEX IF NOT EXISTS transactions_hash_idx ON " + transactionsTable + " USING hash(hash);"
	txMemoIndex := "CREATE INDEX IF NOT EXISTS transactions_memo_idx ON " + transactionsTable + " USING hash(memo) WHERE memo IS NOT NULL;"
	accountTxsIndex := "CREATE INDEX IF NOT EXISTS account_transactions_address_idx ON " + accountTransactionsTable + " USING hash(address);"
	commitSigsIndex := "CREATE INDEX IF NOT EXISTS commit_signatures_block_idx ON " + commitSignaturesTable + " USING hash(block_id);"

	_, err := p.exec.ExecContext(ctx, blockPK)
	if err != nil {
		return errors.New(err, "Create blocks primary key")
	}

	_, err = p.exec.ExecContext(ctx, blockHeightIndex)
	if err != nil {
		return errors.New(err, "Create blocks height index")
	}

	_, err = p.exec.ExecContext(ctx, txFK)
	if err != nil {
		return errors.New(err, "Create transactions foreign index")
	}

	_, err = p.exec.ExecContext(ctx, txBlockIDIndex)
	if err != nil {
		return errors.New(err, "Create transactions block id index")
	}

	_, err = p.exec.ExecContext(ctx, txHashIndex)
	if err != nil {
		return errors.New(err, "Create transactions hash index")
	}

	_, err = p.exec.ExecContext(ctx, txMemoIndex)
	if err != nil {
		return errors.New(err, "Create transactions memo index")
	}

	_, err = p.exec.ExecContext(ctx, accountTxsIndex)
	if err != nil {
		return errors.New(err, "Create account transactions address index")
	}

	_, err = p.exec.ExecContext(ctx, commitSigsIndex)
	return errors.New(err, "Create commit signatures block index")
}

func (p *postgres) ExecContext(ctx context.Context, query string, args ...any) error {
	_, err := p.exec.ExecContext(ctx, query, args...)
	return err
}

func (p *postgres) RunInTransaction(ctx context.Context, txFunc repository.InTransaction) (err error) {
	if p.db == nil {
		return errors.Create("Can't run transaction inside of another transaction")
	}

	var tx *sql.Tx
	tx, err = p.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.New(err, "Begin db tx")
	}

	defer func() {
		if p := recover(); p != nil {
			rErr := tx.Rollback()
			logger.Error("Database rollback failed before panic", zap.Error(rErr))
			panic(p)
		} else if err != nil {
			rErr := tx.Rollback()
			if rErr != nil {
				err = errors.New(err, rErr.Error())
			}
		} else {
			err = tx.Commit()
		}
	}()

	runner := &postgres{p.config, nil, tx, p.psql}
	err = txFunc(ctx, runner)

	return
}
