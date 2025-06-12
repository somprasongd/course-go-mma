package eventhandler

import (
	"context"
	"go-mma/modules/customer/internal/domain/event"
	notiService "go-mma/modules/notification/service"
	"go-mma/shared/common/domain"
)

// customerCreatedDomainEventHandler คือ handler สำหรับจัดการ event ประเภท CustomerCreatedDomainEvent
type customerCreatedDomainEventHandler struct {
	notiSvc notiService.NotificationService // service สำหรับส่งการแจ้งเตือน (เช่น อีเมล)
}

// NewCustomerCreatedDomainEventHandler คือฟังก์ชันสร้าง instance ของ handler นี้
func NewCustomerCreatedDomainEventHandler(notiSvc notiService.NotificationService) domain.DomainEventHandler {
	return &customerCreatedDomainEventHandler{
		notiSvc: notiSvc,
	}
}

// Handle คือฟังก์ชันหลักที่ถูกเรียกเมื่อมี event ถูก dispatch มา
func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
	// แปลง (type assert) event ที่รับมาเป็น pointer ของ CustomerCreatedDomainEvent
	e, ok := evt.(*event.CustomerCreatedDomainEvent)
	if !ok {
		// ถ้าไม่ใช่ event ประเภทนี้ ให้ส่ง error กลับไป
		return domain.ErrInvalidEvent
	}

	// เรียกใช้ service ส่งอีเมลต้อนรับลูกค้าใหม่
	if err := h.notiSvc.SendEmail(e.Email, "Welcome to our service!", map[string]any{
		"message": "Thank you for joining us! We are excited to have you as a member.",
	}); err != nil {
		// หากส่งอีเมลไม่สำเร็จ ส่ง error กลับไป
		return err
	}

	// ถ้าสำเร็จทั้งหมด ให้คืน nil (ไม่มี error)
	return nil
}
