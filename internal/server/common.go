package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/the-laziest/namadexer-go/internal/service"
	"github.com/the-laziest/namadexer-go/pkg/logger"
	"go.uber.org/zap"
)

func (s *Server) getQueryInt64(r *http.Request, name string) int64 {
	value := r.URL.Query().Get(name)
	i64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1
	}
	return i64
}

func (s *Server) getPathInt64(r *http.Request, name string) int64 {
	value, ok := mux.Vars(r)[name]
	if !ok {
		return -1
	}
	i64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1
	}
	return i64
}

func (s *Server) getPathString(r *http.Request, name string) string {
	value, ok := mux.Vars(r)[name]
	if !ok {
		return ""
	}
	return value
}

func (s *Server) getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	status := http.StatusInternalServerError
	switch err {
	case service.ErrBadRequest:
		status = http.StatusBadRequest
	case service.ErrNotFound:
		status = http.StatusNotFound
	}

	return status
}

type responseWriter struct {
	http.ResponseWriter
	response struct {
		result interface{}
		err    error
		code   int
	}
}

func (s *Server) writeResult(w http.ResponseWriter, result interface{}, err error) {
	statusCode := s.getStatusCode(err)

	if rw, ok := w.(*responseWriter); ok {
		rw.response.result = result
		rw.response.err = err
		rw.response.code = statusCode
	}

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(result); err != nil {
			logger.Error("Failed to encode response to JSON", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(statusCode)
}
