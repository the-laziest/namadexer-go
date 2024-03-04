package service

import (
	"context"
	"encoding/json"
)

func (s *service) GetVoteProposalData(ctx context.Context, proposalID int64) ([]json.RawMessage, error) {
	return s.repo.GetVoteProposalDatas(ctx, s.checksums["tx_vote_proposal"], proposalID)
}
