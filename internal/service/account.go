package service

import (
	"context"

	"github.com/the-laziest/namadexer-go/internal/repository"
)

func (s *service) GetAccountUpdates(ctx context.Context, accountID string) (*AccountUpdates, error) {
	thresholds, err := s.repo.GetAccountThresholds(ctx, s.checksums["tx_update_account"], accountID)
	if err == repository.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	vpCodes, err := s.repo.GetAccountVPCodes(ctx, s.checksums["tx_update_account"], accountID)
	if err == repository.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	publicKeys, err := s.repo.GetAccountPublicKeys(ctx, s.checksums["tx_update_account"], accountID)
	if err != nil {
		return nil, err
	}

	return &AccountUpdates{
		AccountID:  accountID,
		CodeHashes: vpCodes,
		Thresholds: thresholds,
		PublicKeys: publicKeys,
	}, nil
}
