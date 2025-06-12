package domain

// Aggregate เป็น struct พื้นฐานสำหรับ aggregate root ทั้งหมดใน DDD
// ใช้เพื่อเก็บรวบรวม domain events ที่เกิดขึ้นภายใน aggregate
type Aggregate struct {
	domainEvents []DomainEvent // เก็บรายการของ event ที่เกิดขึ้นใน aggregate นี้
}

// AddDomainEvent ใช้สำหรับเพิ่ม domain event เข้าไปใน aggregate
// ฟังก์ชันนี้จะถูกเรียกภายใน method อื่น ๆ ของ aggregate เมื่อต้องการประกาศว่า event บางอย่างได้เกิดขึ้นแล้ว
func (a *Aggregate) AddDomainEvent(dv DomainEvent) {
	// สร้าง slice เปล่าหากยังไม่มี
	if a.domainEvents == nil {
		a.domainEvents = make([]DomainEvent, 0)
	}

	// เพิ่ม event ลงใน slice
	a.domainEvents = append(a.domainEvents, dv)
}

// PullDomainEvents จะดึง domain events ทั้งหมดออกจาก aggregate
// พร้อมกับเคลียร์ events เหล่านั้นจาก memory (เพราะ events ถูกส่งออกไปแล้ว)
// เหมาะสำหรับใช้ใน layer ที่ทำการ publish หรือ persist event
func (a *Aggregate) PullDomainEvents() []DomainEvent {
	events := a.domainEvents // ดึง event ทั้งหมดที่บันทึกไว้
	a.domainEvents = nil     // เคลียร์ event list เพื่อป้องกันการส่งซ้ำ
	return events
}
