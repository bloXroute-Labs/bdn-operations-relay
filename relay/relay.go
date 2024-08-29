package relay

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/bloXroute-Labs/bdn-operations-relay/config"
	"golang.org/x/sync/errgroup"
)

func Run(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	eg, gCtx := errgroup.WithContext(ctx)

	return eg.Wait()
}
