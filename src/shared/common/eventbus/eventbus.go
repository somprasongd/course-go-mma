package eventbus

import (
	"context"
)

// IntegrationEventHandler คือ interface สำหรับ handler ที่จะรับผิดชอบการจัดการ Integration Event
// จะต้องมี method Handle เพื่อรับ context และ event ที่ต้องการจัดการ
type IntegrationEventHandler interface {
	Handle(ctx context.Context, event Event) error
}

// EventBus คือ interface สำหรับระบบ event bus ที่ใช้ในการ publish และ subscribe event ต่างๆ
type EventBus interface {
	// Publish ใช้สำหรับส่ง event ออกไปยังระบบ event bus
	// โดยรับ context และ event ที่จะส่ง
	Publish(ctx context.Context, event Event) error

	// Subscribe ใช้สำหรับลงทะเบียน handler สำหรับ event ที่มีชื่อ eventName
	// เมื่อมี event ที่ตรงกับชื่อ eventName เข้ามา handler ที่ลงทะเบียนไว้จะถูกเรียกใช้
	Subscribe(eventName EventName, handler IntegrationEventHandler)
}
