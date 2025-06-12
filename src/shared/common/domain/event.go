package domain

import "time"

// EventName คือ alias ของ string เพื่อใช้แทนชื่อ event เช่น "CustomerCreated", "OrderPlaced" เป็นต้น
type EventName string

// DomainEvent เป็น interface สำหรับ event ที่เกิดขึ้นใน domain (Domain Event ตาม DDD)
// ใช้เพื่อให้สามารถบันทึกหรือส่ง event ได้ โดยไม่ต้องรู้โครงสร้างภายใน
type DomainEvent interface {
	EventName() EventName  // คืนชื่อ event
	OccurredAt() time.Time // คืนเวลาที่ event เกิด
}

// BaseDomainEvent เป็น struct พื้นฐานที่ implement DomainEvent
// ใช้ฝังใน struct อื่นๆ ที่เป็น event เพื่อ reuse method ได้
type BaseDomainEvent struct {
	Name EventName // ชื่อของ event เช่น "UserRegistered"
	At   time.Time // เวลาที่ event นี้เกิดขึ้น
}

// EventName คืนชื่อของ event นี้
func (e BaseDomainEvent) EventName() EventName {
	return e.Name
}

// OccurredAt คืนเวลาที่ event นี้เกิดขึ้น
// ใช้แบบ value receiver ด้วยเหตุผลเดียวกันกับข้างต้น
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.At
}
