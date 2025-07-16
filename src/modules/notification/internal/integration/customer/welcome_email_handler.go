package customer

import (
	"context"
	"fmt"
	"go-mma/modules/notification/service"
	"go-mma/shared/common/eventbus"
	"go-mma/shared/messaging"

	"go.opentelemetry.io/otel/trace"
)

type welcomeEmailHandler struct {
	notiService service.NotificationService
}

func NewWelcomeEmailHandler(notiService service.NotificationService) *welcomeEmailHandler {
	return &welcomeEmailHandler{
		notiService: notiService,
	}
}

func (h *welcomeEmailHandler) Handle(ctx context.Context, evt eventbus.Event) error {
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("integration_event")
	ctx, span := tracer.Start(ctx, "IntegrationEvent:WelcomeEmail")
	defer span.End()

	e, ok := evt.(*messaging.CustomerCreatedIntegrationEvent) // ใช้ pointer
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	return h.notiService.SendEmail(ctx, e.Email, "Welcome to our service!", map[string]any{
		"message": "Thank you for joining us! We are excited to have you as a member.",
	})
}
