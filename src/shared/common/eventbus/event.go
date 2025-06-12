package eventbus

import (
	"time"
)

// EventName เป็นชนิดข้อมูลสำหรับชื่อ event
type EventName string

// Event คือ interface สำหรับ event ทั่วไปในระบบ
// ต้องมี method สำหรับดึงข้อมูล ID, ชื่อ event, และเวลาที่เกิด event
type Event interface {
	EventID() string       // คืนค่า ID ของ event (เช่น UUID หรือ ULID)
	EventName() EventName  // คืนค่าชื่อ event เช่น "CustomerCreated"
	OccurredAt() time.Time // เวลาที่ event นั้นเกิดขึ้น
}

// BaseEvent คือ struct พื้นฐานที่ใช้เก็บข้อมูล event ทั่วไป
// สามารถนำไปฝังใน struct event ที่เฉพาะเจาะจงได้
type BaseEvent struct {
	ID   string    // รหัส event แบบ unique (UUID/ULID)
	Name EventName // ชื่อของ event เช่น "CustomerCreated"
	At   time.Time // เวลาที่ event เกิดขึ้น
}

// EventID คืนค่า ID ของ event
func (e BaseEvent) EventID() string {
	return e.ID
}

// EventName คืนค่าชื่อของ event
func (e BaseEvent) EventName() EventName {
	return e.Name
}

// OccurredAt คืนค่าเวลาที่ event เกิดขึ้น
func (e BaseEvent) OccurredAt() time.Time {
	return e.At
}
