package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
)

type route struct {
	name        string
	method      string
	pattern     string
	handlerFunc http.HandlerFunc
}

func (s *Server) setupHandlers() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	log := func(inner http.Handler, name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			w.Header().Set("Access-Control-Allow-Origin", "*")
			inner.ServeHTTP(w, r)
			logger.Info(fmt.Sprintf("served %s", name), "method", r.Method, "url", r.RequestURI, "duration", time.Since(start))
		})
	}

	for _, r := range s.buildRoutes() {
		var handler http.Handler = r.handlerFunc
		handler = log(handler, r.name)

		router.Methods(r.method).
			Path(r.pattern).
			Name(r.name).
			Handler(handler)
	}

	return router
}

func (s *Server) buildRoutes() []route {
	routes := []route{
		{
			name:        "Ping",
			method:      http.MethodGet,
			pattern:     "/ping",
			handlerFunc: s.ping,
		},
	}

	if s.cfg.DAppPrivateKey != "" {
		routes = append(routes, s.dAppRoutes()...)
	}

	if s.cfg.SolverPrivateKey != "" {
		routes = append(routes, s.solverRoutes()...)
	}

	return routes
}

func (s *Server) dAppRoutes() []route {
	return []route{
		{
			name:        "SubmitUserOperation",
			method:      http.MethodPost,
			pattern:     "/userOperation",
			handlerFunc: s.userOperation,
		},
		{
			name:        "GetSolverOperations",
			method:      http.MethodGet,
			pattern:     "/solverOperations",
			handlerFunc: s.solverOperations,
		},
	}
}

func (s *Server) solverRoutes() []route {
	return []route{{
		"WebsocketSolver",
		http.MethodGet,
		"/ws/solver",
		s.websocketSolver,
	}}
}
