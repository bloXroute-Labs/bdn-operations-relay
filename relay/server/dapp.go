package server

import (
	"net/http"

	"github.com/FastLane-Labs/atlas-operations-relay/operation"
	"github.com/ethereum/go-ethereum/common"
)

func (s *Server) userOperation(w http.ResponseWriter, r *http.Request) {
	var userOperation *operation.UserOperationWithHintsRaw
	err := parseRequest(r, &userOperation)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	userOp, hints := userOperation.Decode()

	partialOperation := operation.NewUserOperationPartialRaw(common.Hash{}, userOp, hints) // TODO hash

	intentID, err := s.intentService.SubmitIntent(r.Context(), partialOperation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeResponseData(w, `{"intent_id": "`+intentID+`"}`)
}

type solverOperationsRequest struct {
	IntentID string `json:"intent_id"`
}

func (s *Server) solverOperations(w http.ResponseWriter, r *http.Request) {
	var req solverOperationsRequest
	err := parseRequest(r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	resp, err := s.intentService.GetIntentSolutions(r.Context(), req.IntentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeResponseData(w, resp)
}
