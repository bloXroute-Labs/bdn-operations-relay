package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bloXroute-Labs/bdn-operations-relay/log"
	"github.com/gorilla/mux"
)

type route struct {
	name        string
	method      string
	pattern     string
	handlerFunc http.HandlerFunc
}

func (s *Server) setupHandlers() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	logger := func(inner http.Handler, name string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			w.Header().Set("Access-Control-Allow-Origin", "*")
			inner.ServeHTTP(w, r)
			log.Info(fmt.Sprintf("served %s", name), "method", r.Method, "url", r.RequestURI, "duration", time.Since(start))
		})
	}

	for _, r := range buildRoutes(s) {
		var handler http.Handler = r.handlerFunc
		handler = logger(handler, r.name)

		router.Methods(r.method).
			Path(r.pattern).
			Name(r.name).
			Handler(handler)
	}

	return router
}

func buildRoutes(s *Server) []route {
	return []route{
		{
			name:        "Ping",
			method:      http.MethodGet,
			pattern:     "/ping",
			handlerFunc: s.ping,
		},
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
