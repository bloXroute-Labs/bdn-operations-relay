package server

import (
	"net/http"

	"github.com/FastLane-Labs/atlas-operations-relay/core"
	"github.com/FastLane-Labs/atlas-operations-relay/operation"
	"github.com/ethereum/go-ethereum/common"
)

func (s *Server) userOperation(w http.ResponseWriter, r *http.Request) {
	var userOperation *operation.UserOperationWithHintsRaw
	if relayErr := parseRequest(r, &userOperation); relayErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(relayErr.Marshal())
		return
	}

	userOp, hints := userOperation.Decode()

	partialOperation := operation.NewUserOperationPartialRaw(common.Hash{}, userOp, hints) // TODO hash

	err := s.intentService.SubmitIntent(r.Context(), partialOperation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(core.ErrI
		return
	}

}

func (s *Server) solverOperations(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
