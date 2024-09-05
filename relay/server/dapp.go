package server

import (
	"net/http"

	"github.com/FastLane-Labs/atlas-sdk-go/types"
)

func (s *Server) userOperation(w http.ResponseWriter, r *http.Request) {
	var req *types.UserOperationWithHintsRaw
	err := parseRequest(r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	chainID, userOp, hints := req.Decode()
	partialOperation, err := types.NewUserOperationPartialRaw(chainID, userOp, hints)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	intentID, err := s.intentService.SubmitIntent(r.Context(), partialOperation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeResponseData(w, map[string]string{
		"intent_id": intentID,
	})
}

func (s *Server) solverOperations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	intentID := q.Get("intentID")
	if intentID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("intentID is required"))
		return
	}

	resp, err := s.intentService.GetIntentSolutions(r.Context(), intentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeResponseData(w, resp)
}
