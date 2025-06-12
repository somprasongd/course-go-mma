package eventbus

import (
	"context"
	"log"
	"sync"
)

// implementation ของ EventBus แบบง่าย ๆ ที่เก็บ subscriber ไว้ใน memory
type inmemoryEventBus struct {
	subscribers map[EventName][]IntegrationEventHandler // เก็บ eventName กับ list ของ handler
	mu          sync.RWMutex                            // mutex สำหรับป้องกัน concurrent access
}

// สร้าง instance ใหม่ของ inmemoryEventBus พร้อม map subscribers ว่าง ๆ
func NewInMemoryEventBus() EventBus {
	return &inmemoryEventBus{
		subscribers: make(map[EventName][]IntegrationEventHandler),
	}
}

// Subscribe ใช้ลงทะเบียน handler สำหรับ event ที่มีชื่อ eventName
// โดยจะเพิ่ม handler เข้าไปใน map subscribers
func (eb *inmemoryEventBus) Subscribe(eventName EventName, handler IntegrationEventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// เพิ่ม handler เข้า slice ของ eventName นั้น ๆ
	eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
}

// Publish ส่ง event ไปยัง handler ทุกตัวที่ subscribe event ชื่อเดียวกัน
func (eb *inmemoryEventBus) Publish(ctx context.Context, event Event) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// หา handler ที่ลงทะเบียนกับ event นี้
	handlers, ok := eb.subscribers[event.EventName()]
	if !ok {
		// ไม่มี handler สำหรับ event นี้ ก็ return nil
		return nil
	}

	// สร้าง context ใหม่ที่อาจมีข้อมูลเพิ่มเติมสำหรับ event bus
	busCtx := context.WithValue(ctx, "name", "context in event bus")

	// เรียก handler ทุกตัวแบบ asynchronous (goroutine) เพื่อไม่บล็อกการทำงาน
	for _, handler := range handlers {
		go func(h IntegrationEventHandler) {
			// เรียก handle event และ log error ถ้ามี
			err := h.Handle(busCtx, event)
			if err != nil {
				log.Printf("error handling event %s: %v", event.EventName(), err)
			}
		}(handler)
	}
	return nil
}
