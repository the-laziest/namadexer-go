package service

import (
	"context"

	"github.com/the-laziest/namadexer-go/internal/repository"
)

func (s *service) GetValidatorsUptime(ctx context.Context, validator string, start, end int64) (Uptime, error) {
	address, err := hexToBytes(validator)
	if err != nil {
		return Uptime{}, err
	}

	if start == -1 || end == -1 {
		end, err = s.repo.GetLastHeight(ctx)
		if err != nil {
			return Uptime{}, err
		}
		start = end - 500
		if start < 1 {
			start = 1
		}
	}

	cnt, err := s.repo.GetCommitsCount(ctx, address, start, end)
	if err == repository.ErrNotFound {
		return Uptime{}, ErrNotFound
	}
	if err != nil {
		return Uptime{}, err
	}

	uptime := float64(cnt) / float64(end-start)

	return Uptime{uptime}, nil
}
