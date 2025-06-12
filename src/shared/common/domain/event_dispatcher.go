package domain

import (
	"context"
	"fmt"
	"sync"
)

// Error ที่ใช้ตรวจสอบความถูกต้องของ event
var (
	ErrInvalidEvent = fmt.Errorf("invalid domain event")
)

// DomainEventHandler คือ interface ที่ทุก handler ของ event ต้อง implement
// โดยจะมี method เดียวคือ Handle เพื่อรับ event และทำงานตาม logic ที่ต้องการ
type DomainEventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
}

// DomainEventDispatcher คือ interface สำหรับระบบที่ทำหน้าที่กระจาย (dispatch) event
// โดยสามารถ register handler สำหรับแต่ละ EventName และ dispatch หลาย event พร้อมกันได้
type DomainEventDispatcher interface {
	Register(eventType EventName, handler DomainEventHandler)
	Dispatch(ctx context.Context, events []DomainEvent) error
}

// simpleDomainEventDispatcher เป็น implementation ง่าย ๆ ของ DomainEventDispatcher
// ใช้ map เก็บ handler แยกตาม EventName
type simpleDomainEventDispatcher struct {
	handlers map[EventName][]DomainEventHandler // แผนที่ของ EventName ไปยัง handler หลายตัว
	mu       sync.RWMutex                       // ใช้ mutex เพื่อป้องกัน concurrent read/write
}

// NewSimpleDomainEventDispatcher สร้าง instance ใหม่ของ dispatcher
func NewSimpleDomainEventDispatcher() DomainEventDispatcher {
	return &simpleDomainEventDispatcher{
		handlers: make(map[EventName][]DomainEventHandler),
	}
}

// Register ใช้สำหรับลงทะเบียน handler กับ EventName
// Handler จะถูกเรียกเมื่อมี event นั้น ๆ ถูก dispatch
func (d *simpleDomainEventDispatcher) Register(eventType EventName, handler DomainEventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// เพิ่ม handler ไปยัง slice ของ event นั้น ๆ
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Dispatch รับ slice ของ event แล้ว dispatch ไปยัง handler ที่ลงทะเบียนไว้
// ถ้ามี handler มากกว่าหนึ่งตัวสำหรับ event เดียวกัน จะเรียกทุกตัว
func (d *simpleDomainEventDispatcher) Dispatch(ctx context.Context, events []DomainEvent) error {
	for _, event := range events {
		// อ่าน handler ของ event นี้ (copy slice เพื่อป้องกัน concurrent modification)
		d.mu.RLock()
		handlers := append([]DomainEventHandler(nil), d.handlers[event.EventName()]...)
		d.mu.RUnlock()

		// เรียก handler แต่ละตัว
		for _, handler := range handlers {
			err := func(h DomainEventHandler) error {
				// หาก handler ทำงานผิดพลาด จะคืน error พร้อมระบุ event ที่ผิด
				err := h.Handle(ctx, event)
				if err != nil {
					return fmt.Errorf("error handling event %s: %w", event.EventName(), err)
				}
				return nil
			}(handler)

			// หากมี error จาก handler ใด ๆ จะหยุดและ return เลย
			if err != nil {
				return err
			}
		}
	}

	// ถ้าไม่มี error เลย ส่ง nil กลับ
	return nil
}
