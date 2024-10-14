package service

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FastLane-Labs/atlas-sdk-go/types"
	sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
	"github.com/bloXroute-Labs/bloxroute-sdk-go/connection/ws"
	"github.com/jellydator/ttlcache/v3"
	"github.com/valyala/fastjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
)

// Intent is a service for interacting with the BDN intent network
type Intent struct {
	client              *sdk.Client
	cfg                 *config.Config
	subscriptionManager *SubscriptionManager
	cache               *ttlcache.Cache[string, []types.SolverOperationRaw]
}

// NewIntent creates a new Intent service
func NewIntent(ctx context.Context, cfg *config.Config, subscriptionManager *SubscriptionManager) (*Intent, error) {
	sdkConfig := &sdk.Config{
		AuthHeader: cfg.BDN.AuthHeader,
		Logger:     new(logger.Instance),
	}

	if cfg.BDN.GRPCURL != "" {
		sdkConfig.GRPCDialOptions = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		}
		sdkConfig.GRPCGatewayURL = cfg.BDN.GRPCURL
	} else {
		sdkConfig.WSDialOptions = &ws.DialOptions{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			HandshakeTimeout: time.Minute,
		}
		sdkConfig.WSGatewayURL = cfg.BDN.WSURL
	}

	client, err := sdk.NewClient(ctx, sdkConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create BDN client: %w", err)
	}

	cache := ttlcache.New[string, []types.SolverOperationRaw](
		ttlcache.WithTTL[string, []types.SolverOperationRaw](time.Minute),
	)

	go cache.Start()

	return &Intent{
		client:              client,
		cfg:                 cfg,
		subscriptionManager: subscriptionManager,
		cache:               cache,
	}, nil
}

// Close closes the connection to the BDN
func (i *Intent) Close() error {
	return i.client.Close()
}

// SubmitIntent submits an intent to the BDN
func (i *Intent) SubmitIntent(ctx context.Context, intent []byte) (string, error) {
	params := &sdk.SubmitIntentParams{
		DappAddress:      i.cfg.DAppAddress,
		SenderPrivateKey: i.cfg.DAppPrivateKey,
		Intent:           intent,
	}

	resp, err := i.client.SubmitIntent(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to submit intent: %w", err)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(*resp)
	if err != nil {
		return "", fmt.Errorf("failed to parse message: %w", err)
	}

	return string(v.GetStringBytes("intent_id")), nil
}

func (i *Intent) SubscribeToIntents(ctx context.Context) error {
	logger.Debug("subscribing to intents")

	params := &sdk.IntentsParams{
		SolverPrivateKey: i.cfg.SolverPrivateKey,
		DappAddress:      i.cfg.DAppAddress,
	}

	err := i.client.OnIntents(ctx, params, func(ctx context.Context, err error, result *sdk.OnIntentsNotification) {
		if err != nil {
			logger.Error("error receiving intent", "error", err)
			return
		}

		logger.Debug("received intent", "dapp_address", result.DappAddress, "sender_address", result.SenderAddress,
			"intent_id", result.IntentID)

		rawIntent := make([]byte, base64.StdEncoding.DecodedLen(len(result.Intent)))
		_, err = base64.StdEncoding.Decode(rawIntent, result.Intent)
		if err == nil {
			result.Intent = rawIntent
		}

		i.subscriptionManager.Notify(result)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to intents: %w", err)
	}

	return nil
}

// SubmitIntentSolution submits an intent solution to the BDN
func (i *Intent) SubmitIntentSolution(ctx context.Context, intentID string, intent []byte) error {
	params := &sdk.SubmitIntentSolutionParams{
		SolverPrivateKey: i.cfg.SolverPrivateKey,
		IntentID:         intentID,
		IntentSolution:   intent,
	}

	_, err := i.client.SubmitIntentSolution(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to submit intent solution: %w", err)
	}

	return nil
}

// GetIntentSolutions gets list of solutions for a specific intent
func (i *Intent) GetIntentSolutions(ctx context.Context, intentID string) ([]types.SolverOperationRaw, error) {
	// check if we have the solutions in cache
	item := i.cache.Get(intentID)
	if item != nil && len(item.Value()) != 0 {
		logger.Debug("returning cached intent solutions", "intent_id", intentID)
		return item.Value(), nil
	}

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

	var result []types.SolverOperationRaw

	for _, obj := range v.GetArray() {
		intentSolution := obj.Get("intent_solution").GetStringBytes()

		out := make([]byte, base64.StdEncoding.DecodedLen(len(intentSolution)))
		n, err := base64.StdEncoding.Decode(out, intentSolution)
		if err != nil {
			logger.Error("failed to decode intent solution from base64", "error", err, "intent_solution", string(intentSolution))
			continue
		}

		var solverOperation *types.SolverOperationRaw
		err = json.Unmarshal(out[:n], &solverOperation) // TODO use var p fastjson.Parser
		if err != nil {
			logger.Error("failed to unmarshal intent solution into SolverOperationRaw", "error", err,
				"intent_solution", string(intentSolution))
			continue
		}

		result = append(result, *solverOperation)
	}

	return result, nil
}

func (i *Intent) SubscribeToSolutions(ctx context.Context) error {
	logger.Debug("subscribing to intent solutions")

	params := &sdk.IntentSolutionsParams{
		DappPrivateKey: i.cfg.DAppPrivateKey,
	}

	err := i.client.OnIntentSolutions(ctx, params, func(ctx context.Context, err error, result *sdk.OnIntentSolutionsNotification) {
		if err != nil {
			logger.Error("error receiving intent solution", "error", err)
			return
		}

		logger.Debug("received intent solution", "intent_id", result.IntentID)

		item := i.cache.Get(result.IntentID)
		if item == nil {
			return
		}

		v := item.Value()

		out := make([]byte, base64.StdEncoding.DecodedLen(len(result.IntentSolution)))
		n, err := base64.StdEncoding.Decode(out, result.IntentSolution)
		if err != nil {
			logger.Error("failed to decode intent solution from base64", "error", err, "intent_solution", string(result.IntentSolution))
			return
		}

		var solverOperation *types.SolverOperationRaw
		err = json.Unmarshal(out[:n], &solverOperation)
		if err != nil {
			logger.Error("failed to unmarshal intent solution into SolverOperationRaw", "error", err,
				"intent_solution", string(result.IntentSolution))
			return
		}

		v = append(v, *solverOperation)

		i.cache.Set(result.IntentID, v, ttlcache.DefaultTTL)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to intent solutions: %w", err)
	}

	return nil
}

func (i *Intent) SubscribeToIntentSolutions(intentID string) {
	i.cache.Set(intentID, []types.SolverOperationRaw{}, ttlcache.DefaultTTL)
}
