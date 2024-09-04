package server

import "net/http"

func (s *Server) ping(w http.ResponseWriter, _ *http.Request) {
	writeResponseData(w, "pong")
}
