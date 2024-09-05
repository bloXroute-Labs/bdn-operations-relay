package server

import (
	"net/http"

	"github.com/FastLane-Labs/atlas-sdk-go/types"
)

type userOperationRequest struct {
	UserOperation *types.UserOperationWithHintsRaw `json:"userOperation"`
}

func (s *Server) userOperation(w http.ResponseWriter, r *http.Request) {
	var req userOperationRequest
	err := parseRequest(r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	chainID, userOp, hints := req.UserOperation.Decode()
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

	writeResponseData(w, `{"intent_id": "`+intentID+`"}`)
}

func (s *Server) solverOperations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	resp, err := s.intentService.GetIntentSolutions(r.Context(), q.Get("intentID"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeResponseData(w, resp)
}
