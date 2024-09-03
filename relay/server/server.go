package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
	"github.com/bloXroute-Labs/bdn-operations-relay/relay/service"
)

// Server handler http calls
type Server struct {
	server        *http.Server
	cfg           *config.Config
	intentService *service.Intent
}

// NewServer creates and returns a new websocket server managed by feedManager
func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	intentService, err := service.NewIntent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create intent service: %v", err)
	}

	return &Server{
		cfg:           cfg,
		intentService: intentService,
	}, nil
}

// Start setup handlers and start http server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%v", s.cfg.HTTPPort),
		ReadHeaderTimeout: time.Second * 5,
	}

	logger.Info("starting HTTP server", "address", s.server.Addr)
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
		logger.Warn("stopping http server that was not initialized")
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("failed to shutdown http server", "error", err)
	}
}

func writeResponseData(w http.ResponseWriter, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		logger.Error("failed to marshal response data", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			logger.Error("failed to write response", "error", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(b)
	if err != nil {
		logger.Error("failed to write response", "error", err)
	}
}

func parseRequest(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal request: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(v); err != nil {
		return fmt.Errorf("invalid request: %v", err)
	}

	return nil
}
