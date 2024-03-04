package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/the-laziest/namadexer-go/pkg/errors"
	"github.com/the-laziest/namadexer-go/pkg/logger"
	"go.uber.org/zap"
)

func (s *Server) logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultFields := []zap.Field{
			zap.String("url", r.URL.String()),
			zap.String("method", r.Method),
		}

		logger.Info("Request", defaultFields...)

		writer := &responseWriter{ResponseWriter: w}

		h.ServeHTTP(writer, r)

		bs, _ := json.Marshal(writer.response.result)

		responseFields := defaultFields
		responseFields = append(responseFields,
			zap.Int("status_code", writer.response.code),
			zap.String("body", string(bs)),
			zap.Error(writer.response.err),
		)

		logger.Info("Response", responseFields...)
	})
}

func (s *Server) recovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				s.writeResult(w, nil, errors.New(fmt.Errorf("%+v", err), "Panic recovered"))
			}
		}()

		h.ServeHTTP(w, r)
	})
}
