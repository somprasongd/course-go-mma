package event

import (
	"go-mma/shared/common/domain"
	"time"
)

// กำหนดชื่อ Event ที่ใช้ในระบบ (EventName)
// เพื่อระบุชนิดของ Domain Event ว่าเป็น "CustomerCreated"
const (
	CustomerCreatedDomainEventType domain.EventName = "CustomerCreated"
)

// CustomerCreatedDomainEvent คือ struct ที่เก็บข้อมูลของเหตุการณ์
// ที่เกิดขึ้นเมื่อมีการสร้าง Customer ใหม่ในระบบ
type CustomerCreatedDomainEvent struct {
	domain.BaseDomainEvent        // ฝัง BaseDomainEvent ที่มีชื่อและเวลาเกิด event
	CustomerID             int64  // ID ของ Customer ที่ถูกสร้าง
	Email                  string // Email ของ Customer ที่ถูกสร้าง
}

// NewCustomerCreatedDomainEvent สร้าง instance ใหม่ของ CustomerCreatedDomainEvent
// โดยรับ customer ID และ email เป็น input และตั้งชื่อ event กับเวลาปัจจุบันอัตโนมัติ
func NewCustomerCreatedDomainEvent(custID int64, email string) *CustomerCreatedDomainEvent {
	return &CustomerCreatedDomainEvent{
		BaseDomainEvent: domain.BaseDomainEvent{
			Name: CustomerCreatedDomainEventType, // กำหนดชื่อ event
			At:   time.Now(),                     // เวลาเกิด event ณ ปัจจุบัน
		},
		CustomerID: custID, // กำหนด customer ID
		Email:      email,  // กำหนด email
	}
}
