package auth_transport_http

import (
	"context"
	core_logger "discord/internal/core/logger"
	"time"
)

const sessionCookieTTL = 24 * time.Hour

func getLogger(ctx context.Context) *core_logger.Logger {
	log := core_logger.FromContext(ctx)
	if log == nil {
		log = core_logger.NewLogger()
	}
	return log
}
