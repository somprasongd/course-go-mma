package eventhandler

import (
	"context"
	"go-mma/modules/customer/internal/domain/event"
	"go-mma/shared/common/domain"
	"go-mma/shared/common/eventbus"
	"go-mma/shared/messaging"

	"go.opentelemetry.io/otel/trace"
)

// customerCreatedDomainEventHandler คือ handler สำหรับจัดการ event ประเภท CustomerCreatedDomainEvent
type customerCreatedDomainEventHandler struct {
	eventBus eventbus.EventBus // เปลี่ยนมาใช้ eventbus
}

// NewCustomerCreatedDomainEventHandler คือฟังก์ชันสร้าง instance ของ handler นี้
func NewCustomerCreatedDomainEventHandler(eventBus eventbus.EventBus) domain.DomainEventHandler {
	return &customerCreatedDomainEventHandler{
		eventBus: eventBus, // เปลี่ยนมาใช้ eventbus
	}
}

// Handle คือฟังก์ชันหลักที่ถูกเรียกเมื่อมี event ถูก dispatch มา
func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("domain_event")
	ctx, span := tracer.Start(ctx, "DomainEvent:CreateCustomer")
	defer span.End()

	// แปลง (type assert) event ที่รับมาเป็น pointer ของ CustomerCreatedDomainEvent
	e, ok := evt.(*event.CustomerCreatedDomainEvent)
	if !ok {
		// ถ้าไม่ใช่ event ประเภทนี้ ให้ส่ง error กลับไป
		return domain.ErrInvalidEvent
	}

	// สร้าง IntegrationEvent จาก Domain Event
	integrationEvent := messaging.NewCustomerCreatedIntegrationEvent(
		e.CustomerID,
		e.Email,
	)

	return h.eventBus.Publish(ctx, integrationEvent)
}
