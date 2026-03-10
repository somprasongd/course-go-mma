## สร้างระบบส่งอีเมลแบบ Reusable ด้วย Notification Service

ในขั้นตอนการสร้างลูกค้าใหม่ เราต้องการเพิ่มฟีเจอร์ **ส่งอีเมลต้อนรับ** หลังจากสร้างลูกค้าสำเร็จ และในอนาคตเราจะใช้ฟีเจอร์ส่งอีเมลในหลายจุด เช่น การยืนยันคำสั่งซื้อ

เพื่อให้สามารถใช้งานซ้ำได้อย่างยืดหยุ่นและง่ายต่อการดูแล เราจะ **แยกการส่งอีเมลออกเป็น Service เฉพาะ** ชื่อว่า `NotificationService` โดยมีขั้นตอนดังนี้

- สร้าง NotificationService แยกออกมา
เราจะสร้าง service ใหม่ในไฟล์ `service/notification.go` เพื่อจัดการการส่งอีเมลทั้งหมดของระบบไว้ในจุดเดียว

    ```go
    package service
    
    import (
     "fmt"
     "go-mma/util/logger"
    )
    
    type NotificationService struct {
    }
    
    func NewNotificationService() *NotificationService {
     return &NotificationService{}
    }
    
    func (s *NotificationService) SendEmail(to string, subject string, payload map[string]any) error {
     // implement email sending logic here
     logger.Log.Info(fmt.Sprintf("Sending email to %s with subject: %s and payload: %v", to, subject, payload))
     return nil
    }
    
    ```

- Inject NotificationService เข้าไปใน CustomerService
แก้ไข `CustomerService` ที่ไฟล์ `service/customer.go` ให้สามารถเรียกใช้ `NotificationService` ได้ เพื่อทำการส่งอีเมลต้อนรับหลังจากสร้างลูกค้าใหม่

    ```go
    package service
    
    // ...
    
    type CustomerService struct {
     custRepo *repository.CustomerRepository
     notiSvc  *NotificationService // <-- เพิ่มตรงนี้
    }
    
    func NewCustomerService(custRepo *repository.CustomerRepository, 
     notiSvc *NotificationService, // <-- เพิ่มตรงนี้
     ) *CustomerService {
     return &CustomerService{
      custRepo: custRepo,
      notiSvc:  notiSvc, // <-- เพิ่มตรงนี้
     }
    }
    
    func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // Rule: email ต้องไม่ซ้ำในฐานข้อมูล
     // แปลง DTO → Model
     // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
    
      // <-- เพิ่มตรงนี้
     // ส่งอีเมลต้อนรับ 
     if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
      "message": "Thank you for joining us! We are excited to have you as a member.",
     }); err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง DTO Response
    }
    ```

- Wiring Dependencies ใน Layer ของ Application
สุดท้าย ต้องแก้ไฟล์ `application/http.go` เพื่อทำ dependency injection ให้กับ `NotificationService` และ `CustomerService` อย่างถูกต้องในจุดเริ่มต้นของแอป

    ```go
    func (s *httpServer) RegisterRoutes(db sqldb.DBContext) {
     v1 := s.app.Group("/api/v1")
    
     customers := v1.Group("/customers")
     {
      repo := repository.NewCustomerRepository(db)
      svcNoti := service.NewNotificationService()      // <-- เพิ่มตรงนี้
      svc := service.NewCustomerService(repo, svcNoti) // <-- เพิ่มตรงนี้
      hdlr := handler.NewCustomerHandler(svc)
      customers.Post("", hdlr.CreateCustomer)
     }
     // orders
    }
    ```

เมื่อแยก notification logic ออกมาเป็น service จะช่วยให้

- Code **แยกความรับผิดชอบ (separation of concerns)** ชัดเจน
- สามารถนำไปใช้ซ้ำใน module อื่น ๆ ได้ทันที เช่น order
- รองรับการเปลี่ยนแปลง เช่น เปลี่ยนจาก log → SMTP → SendGrid โดยไม่กระทบ business logic

