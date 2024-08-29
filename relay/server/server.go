package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/FastLane-Labs/atlas-operations-relay/core"
	"github.com/bloXroute-Labs/bdn-operations-relay/log"
)

// Server handler http calls
type Server struct {
	server *http.Server
	port   int
}

// NewServer creates and returns a new websocket server managed by feedManager
func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start setup handlers and start http server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%v", s.port),
		ReadHeaderTimeout: time.Second * 5,
	}

	log.Info("starting HTTP server", "address", s.server.Addr)
	s.server.Handler = s.setupHandlers()

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start HTTP RPC server: %v", err)
	}

	return nil
}

// Shutdown stops the HTTP server
func (s *Server) Shutdown() {
	if s.server == nil {
		log.Warn("stopping http server that was not initialized")
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("failed to shutdown http server", "error", err)
	}
}

func writeResponseData(w http.ResponseWriter, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Error("failed to marshal response data", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(core.ErrServerCorruptedData.AddError(err).Marshal())
		if err != nil {
			log.Error("failed to write response", "error", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(b)
	if err != nil {
		log.Error("failed to write response", "error", err)
	}
}
