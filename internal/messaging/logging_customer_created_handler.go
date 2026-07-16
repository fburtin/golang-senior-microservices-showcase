package messaging

import (
	"context"
	"log/slog"
)

type LoggingCustomerCreatedHandler struct {
	logger *slog.Logger
}

func NewLoggingCustomerCreatedHandler(logger *slog.Logger) *LoggingCustomerCreatedHandler {
	return &LoggingCustomerCreatedHandler{logger: logger}
}

func (h *LoggingCustomerCreatedHandler) HandleCustomerCreated(
	_ context.Context,
	event CustomerCreatedEvent,
) error {
	h.logger.Info(
		"customer-created event handled",
		"event_id", event.EventID,
		"customer_id", event.ID,
		"email", event.Email,
	)
	return nil
}
