package server

import (
	"net/http"

	"github.com/FastLane-Labs/atlas-sdk-go/types"
	"github.com/ethereum/go-ethereum/log"
)

func (s *Server) userOperation(w http.ResponseWriter, r *http.Request) {
	var userOperation types.UserOperationWithHintsRaw
	err := parseRequest(r, &userOperation)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	chainId, userOp, hints := userOperation.Decode()

	partialOperation, err := types.NewUserOperationPartialRaw(chainId, userOp, hints)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	intentID := r.URL.Query().Get("intent_id")
	if intentID == "" {
		log.Error("missing intent_id")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing intent_id"))
		return
	}

	resp, err := s.intentService.GetIntentSolutions(r.Context(), intentID)
	if err != nil {
		log.Error("failed to get intent solutions", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeResponseData(w, resp)
}
