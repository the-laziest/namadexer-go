package postgres

import (
	"context"

	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

func (p *postgres) AddEvidences(ctx context.Context, evidences ...repository.Evidence) error {
	if len(evidences) == 0 {
		return nil
	}

	builder := p.psql.Insert(evidencesTable).
		Columns("block_id", "height", "time", "address", "total_voting_power", "validator_power")

	for _, evidence := range evidences {
		builder = builder.Values(evidence.BlockID, evidence.Height, evidence.Time, evidence.Address, evidence.TotalVotingPower, evidence.ValidatorPower)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.New(err, "Build SQL for AddEvidences")
	}

	_, err = p.exec.ExecContext(ctx, query, args...)
	return errors.New(err, "Exec SQL for AddEvidences")
}
