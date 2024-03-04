package service

import (
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

type service struct {
	repo      repository.Repository
	checksums map[string][]byte
}

func New(repo repository.Repository, checksums map[string]string) (*service, error) {
	csBytes := make(map[string][]byte, len(checksums))
	for k, v := range checksums {
		vbs, err := hexToBytes(v)
		if err != nil {
			return nil, errors.New(err, "Reformat checksums")
		}
		csBytes[k] = vbs
	}

	if _, ok := csBytes["tx_vote_proposal"]; !ok {
		return nil, errors.Create("tx_vote_proposal hash in checksums not found")
	}
	if _, ok := csBytes["tx_update_account"]; !ok {
		return nil, errors.Create("tx_update_account hash in checksums not found")
	}

	return &service{repo, csBytes}, nil
}

const MASP_ADDR = "tnam1pcqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzmefah"
