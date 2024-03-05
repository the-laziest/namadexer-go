package server

import (
	"net/http"

	"github.com/the-laziest/namadexer-go/internal/service"
)

func (s *Server) blockByHeight(w http.ResponseWriter, r *http.Request) {
	height := s.getPathInt64(r, "height")
	if height == -1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetBlockByHeight(r.Context(), height)

	s.writeResult(w, result, err)
}

func (s *Server) blockByHash(w http.ResponseWriter, r *http.Request) {
	hash := s.getPathString(r, "hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetBlockByHash(r.Context(), hash)

	s.writeResult(w, result, err)
}

func (s *Server) lastBlock(w http.ResponseWriter, r *http.Request) {
	num, offset := s.getQueryInt64(r, "num"), s.getQueryInt64(r, "offset")
	if num <= 0 {
		num = 1
	}

	blocks, err := s.service.GetLatestBlocks(r.Context(), num, offset)
	if err != nil {
		s.writeResult(w, nil, err)
	}

	if num == 1 {
		s.writeResult(w, blocks[0], nil)
	} else {
		s.writeResult(w, blocks, nil)
	}
}

func (s *Server) txsByHashes(w http.ResponseWriter, r *http.Request) {
	hashes, ok := r.URL.Query()["hash"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetTxsByHashes(r.Context(), hashes...)

	s.writeResult(w, result, err)
}

func (s *Server) txByHash(w http.ResponseWriter, r *http.Request) {
	hash := s.getPathString(r, "hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetTxsByHashes(r.Context(), hash)
	if err != nil {
		s.writeResult(w, nil, err)
		return
	}
	if len(result) == 0 {
		s.writeResult(w, nil, service.ErrNotFound)
		return
	}

	s.writeResult(w, result[0], nil)
}

func (s *Server) txsByMemo(w http.ResponseWriter, r *http.Request) {
	memo := s.getPathString(r, "memo")
	if memo == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	limit, offset := s.getQueryInt64(r, "limit"), s.getQueryInt64(r, "offset")

	result, err := s.service.GetTxsByMemo(r.Context(), memo, limit, offset)

	s.writeResult(w, result, err)
}

func (s *Server) txsByMemoTotal(w http.ResponseWriter, r *http.Request) {
	memo := s.getPathString(r, "memo")
	if memo == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetTotalTxsByMemo(r.Context(), memo)

	s.writeResult(w, result, err)
}

func (s *Server) txVoteProposal(w http.ResponseWriter, r *http.Request) {
	proposalID := s.getPathInt64(r, "proposal_id")
	if proposalID == -1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetVoteProposalData(r.Context(), proposalID)

	s.writeResult(w, result, err)
}

func (s *Server) txShielded(w http.ResponseWriter, r *http.Request) {
	result, err := s.service.GetShielded(r.Context())

	s.writeResult(w, result, err)
}

func (s *Server) accountUpdates(w http.ResponseWriter, r *http.Request) {
	accountID := s.getPathString(r, "account_id")
	if accountID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetAccountUpdates(r.Context(), accountID)

	s.writeResult(w, result, err)
}

func (s *Server) accountTxs(w http.ResponseWriter, r *http.Request) {
	address := s.getPathString(r, "account_id")
	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit, offset := s.getQueryInt64(r, "limit"), s.getQueryInt64(r, "offset")

	result, err := s.service.GetTxsByAccount(r.Context(), address, limit, offset)

	s.writeResult(w, result, err)
}

func (s *Server) accountTxsTotal(w http.ResponseWriter, r *http.Request) {
	address := s.getPathString(r, "account_id")
	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.service.GetTotalTxsByAccount(r.Context(), address)

	s.writeResult(w, result, err)
}

func (s *Server) validatorUptime(w http.ResponseWriter, r *http.Request) {
	address := s.getPathString(r, "validator_address")
	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	start, end := s.getQueryInt64(r, "start"), s.getQueryInt64(r, "end")

	result, err := s.service.GetValidatorsUptime(r.Context(), address, start, end)

	s.writeResult(w, result, err)
}
