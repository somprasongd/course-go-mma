package customer

import (
	"context"
	"fmt"
	"go-mma/modules/notification/service"
	"go-mma/shared/common/eventbus"
	"go-mma/shared/messaging"
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
	e, ok := evt.(*messaging.CustomerCreatedIntegrationEvent) // ใช้ pointer
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	return h.notiService.SendEmail(ctx, e.Email, "Welcome to our service!", map[string]any{
		"message": "Thank you for joining us! We are excited to have you as a member.",
	})
}
