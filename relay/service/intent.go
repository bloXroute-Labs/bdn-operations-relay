package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FastLane-Labs/atlas-operations-relay/operation"
	sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
	"github.com/bloXroute-Labs/bloxroute-sdk-go/connection/ws"
	"github.com/valyala/fastjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
)

// Intent is a service for interacting with the BDN intent network
type Intent struct {
	client *sdk.Client
	cfg    *config.Config
}

// NewIntent creates a new Intent service
func NewIntent(ctx context.Context, cfg *config.Config) (*Intent, error) {
	sdkConfig := &sdk.Config{
		AuthHeader:     cfg.BDN.AuthHeader,
		WSGatewayURL:   cfg.BDN.WSURL,
		GRPCGatewayURL: cfg.BDN.GRPCURL,
	}

	if cfg.BDN.GRPCURL != "" {
		sdkConfig.GRPCDialOptions = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		}
	} else {
		sdkConfig.WSDialOptions = &ws.DialOptions{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			HandshakeTimeout: time.Minute,
		}
	}

	client, err := sdk.NewClient(ctx, sdkConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create BDN client: %w", err)
	}

	return &Intent{
		client: client,
		cfg:    cfg,
	}, nil
}

// Close closes the connection to the BDN
func (i *Intent) Close() error {
	return i.client.Close()
}

// SubmitIntent submits an intent to the BDN
func (i *Intent) SubmitIntent(ctx context.Context, userOp *operation.UserOperationPartialRaw) error {
	intent, err := json.Marshal(userOp)
	if err != nil {
		return fmt.Errorf("failed to marshal user operation: %w", err)
	}

	params := &sdk.SubmitIntentParams{
		DappAddress:      i.cfg.DAppAddress,
		SenderPrivateKey: i.cfg.DAppPrivateKey,
		Intent:           intent,
	}

	_, err = i.client.SubmitIntent(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to submit intent: %w", err)
	}

	return nil
}

// SubmitIntentSolution submits an intent solution to the BDN
func (i *Intent) SubmitIntentSolution(ctx context.Context, intentID string, intentSolution []byte) error {
	params := &sdk.SubmitIntentSolutionParams{
		SolverPrivateKey: i.cfg.SolverPrivateKey,
		IntentID:         intentID,
		IntentSolution:   intentSolution,
	}

	_, err := i.client.SubmitIntentSolution(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to submit intent solution: %w", err)
	}

	return nil
}

// GetIntentSolutions gets list of solutions for a specific intent
func (i *Intent) GetIntentSolutions(ctx context.Context, intentID string) ([]operation.SolverOperationRaw, error) {
	params := &sdk.GetSolutionsForIntentParams{
		DAppOrSenderPrivateKey: i.cfg.DAppPrivateKey,
		IntentID:               intentID,
	}

	resp, err := i.client.GetSolutionsForIntent(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get intent solutions: %w", err)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(*resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	var result []operation.SolverOperationRaw

	for _, obj := range v.GetArray() {
		intentSolution := obj.Get("intent_solution").GetStringBytes()

		var solverOperation operation.SolverOperationRaw
		err = json.Unmarshal(intentSolution, &solverOperation) // TODO use var p fastjson.Parser
		if err != nil {
			logger.Error("failed to unmarshal intent solution into SolverOperationRaw", "error", err,
				"intent_solution", string(intentSolution))
			continue
		}

		result = append(result, solverOperation)
	}

	return result, nil
}
