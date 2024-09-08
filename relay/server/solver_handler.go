package server

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/valyala/fastjson"

	"github.com/bloXroute-Labs/bdn-operations-relay/logger"
	"github.com/bloXroute-Labs/bdn-operations-relay/relay/service"
)

const (
	methodPing                  = "ping"
	methodSubscribe             = "subscribe"
	methodUnsubscribe           = "unsubscribe"
	methodSubmitSolverOperation = "submitSolverOperation"

	microSecTimeFormat = "2006-01-02 15:04:05.000000"
)

type wsConnHandler struct {
	remoteAddress       string
	intentService       *service.Intent
	subscriptionService *service.SubscriptionManager
}

func (h *wsConnHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	method := req.Method

	switch method {
	case methodPing:
		response := pingResponse{
			Pong: time.Now().UTC().Format(microSecTimeFormat),
		}
		if err := conn.Reply(ctx, req.ID, response); err != nil {
			logger.Error("error replying to client", "err", err, "reqID", req.ID, "caller", h.remoteAddress)
		}
	case methodSubscribe:
		h.handleSubscribe(ctx, conn, req)
	case methodUnsubscribe:
		h.handleUnsubscribe(ctx, conn, req)
	case methodSubmitSolverOperation:
		h.handleSubmitIntentSolution(ctx, conn, req)
	default:
		h.sendErrorMsg(ctx, jsonrpc2.CodeMethodNotFound, "unsupported method name: "+method, conn, req.ID)
	}
}

// handleSubscribe handles the subscribe method
func (h *wsConnHandler) handleSubscribe(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params == nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidParams, "params value is missing", conn, req.ID)
		return
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(*req.Params)
	if err != nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidParams, fmt.Sprintf("failed to parse params: %v", err), conn, req.ID)
		return
	}

	subscriptionType := v.GetStringBytes("subscription_type")
	subscription, err := h.subscriptionService.Subscribe(h.remoteAddress, service.SubscriptionType(subscriptionType))
	if err != nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidRequest, fmt.Sprintf("failed to subscribe: %v", err), conn, req.ID)
		return
	}

	response := subscribeResponse{
		SubscriptionID: subscription.ID,
	}

	if err = conn.Reply(ctx, req.ID, response); err != nil {
		logger.Error("error replying to client", "err", err, "reqID", req.ID, "caller", h.remoteAddress)
		return
	}

	logger.Info("client subscribed", "subscription_type", string(subscriptionType), "caller", h.remoteAddress)

	h.handlerSubscriptionMessages(ctx, conn, subscription)
}

// handleUnsubscribe handles the unsubscribe method
func (h *wsConnHandler) handleUnsubscribe(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params == nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidParams, "params value is missing", conn, req.ID)
		return
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(*req.Params)
	if err != nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidParams, fmt.Sprintf("failed to parse params: %v", err), conn, req.ID)
		return
	}

	subscriptionID := v.GetStringBytes("subscription_id")
	err = h.subscriptionService.Unsubscribe(h.remoteAddress, string(subscriptionID))
	if err != nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidRequest, fmt.Sprintf("failed to unsubscribe: %v", err), conn, req.ID)
		return
	}

	if err = conn.Reply(ctx, req.ID, "true"); err != nil {
		logger.Error("error replying to client", "err", err, "reqID", req.ID, "caller", h.remoteAddress)
		return
	}

	logger.Info("client unsubscribed", "subscriptionID", string(subscriptionID), "caller", h.remoteAddress)
}

// handleSubmitIntentSolution handles the submitSolverOperation method
func (h *wsConnHandler) handleSubmitIntentSolution(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params == nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidParams, "params value is missing", conn, req.ID)
		return
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(*req.Params)
	if err != nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInvalidParams, fmt.Sprintf("failed to parse params: %v", err), conn, req.ID)
		return
	}

	operationID := v.GetStringBytes("operation_id")
	operation := v.GetObject("operation")

	err = h.intentService.SubmitIntentSolution(ctx, string(operationID), operation.MarshalTo(nil))
	if err != nil {
		h.sendErrorMsg(ctx, jsonrpc2.CodeInternalError, fmt.Sprintf("failed to submit intent solution: %v", err), conn, req.ID)
	}
}

// sendErrorMsg formats and sends an RPC error message back to the client
func (h *wsConnHandler) sendErrorMsg(ctx context.Context, code int, message string, conn *jsonrpc2.Conn, reqID jsonrpc2.ID) {
	rpcError := &jsonrpc2.Error{
		Code:    int64(code),
		Message: message,
	}

	err := conn.ReplyWithError(ctx, reqID, rpcError)
	if err != nil {
		logger.Error("could not respond to client with error message", "err", err, "reqID", reqID, "caller", h.remoteAddress)
	}
}

func (h *wsConnHandler) handlerSubscriptionMessages(ctx context.Context, conn *jsonrpc2.Conn, subscription *service.Subscription) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-subscription.NotificationChannel:
			if !ok {
				return
			}

			err := conn.Notify(ctx, "subscribe", msg)
			if err != nil {
				logger.Error("error replying to client", "err", err, "caller", h.remoteAddress)
				return
			}
		}
	}
}
