package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/the-laziest/namadexer-go/internal/service"
	"github.com/the-laziest/namadexer-go/pkg/logger"
)

type Server struct {
	config  Config
	server  *http.Server
	router  *mux.Router
	service service.Service
}

func New(config Config, service service.Service) *Server {
	router := mux.NewRouter()
	server := &Server{
		config: config,
		server: &http.Server{
			Addr:    ":" + config.Port,
			Handler: router,
		},
		router:  router,
		service: service,
	}
	server.init()
	return server
}

func (s *Server) init() {
	s.router.Use(s.logMiddleware, s.recovery)

	methods := []string{http.MethodGet, http.MethodPost}

	routes := []struct {
		path        string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}{
		{"/block/height/{height:[0-9]+}", s.blockByHeight},
		{"/block/hash/{hash}", s.blockByHash},
		{"/block/last", s.lastBlock},
		{"/txs", s.txsByHashes},
		{"/txs/memo/{memo}", s.txsByMemo},
		{"/txs/memo/{memo}/total", s.txsByMemoTotal},
		{"/tx/vote_proposal/{proposal_id:[0-9]+}", s.txVoteProposal},
		{"/tx/shielded", s.txShielded},
		{"/tx/{hash}", s.txByHash},
		{"/account/updates/{account_id}", s.accountUpdates},
		{"/account/txs/{account_id}", s.accountTxs},
		{"/account/txs/{account_id}/total", s.accountTxsTotal},
		{"/validator/{validator_address}/uptime", s.validatorUptime},
	}

	for _, route := range routes {
		s.router.HandleFunc(route.path, route.handlerFunc).Methods(methods...)
	}
}

func (s *Server) Start() error {
	logger.Info("Starting server", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
