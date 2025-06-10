package service

import (
	"context"
	"go-mma/dto"
	"go-mma/model"
	"go-mma/repository"
	"go-mma/util/errs"
	"go-mma/util/logger"
	"go-mma/util/storage/sqldb/transactor"
)

var (
	ErrCreditValue = errs.BusinessRuleError("credit must be greater than 0")
	ErrEmailExists = errs.ConflictError("email already exists")
)

// --> Step 1: สร้าง interface
type CustomerService interface {
	CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error)
}

// --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
type customerService struct {
	transactor transactor.Transactor
	custRepo   repository.CustomerRepository // --> step 3: เปลี่ยนจาก pointer เป็น interface
	notiSvc    NotificationService           // --> step 4: เปลี่ยนจาก pointer เป็น interface
}

func NewCustomerService(
	transactor transactor.Transactor,
	custRepo repository.CustomerRepository, // --> step 5: เปลี่ยนจาก pointer เป็น interface
	notiSvc NotificationService, // --> step 6: เปลี่ยนจาก pointer เป็น interface
) CustomerService { // --> Step 7: return เป็น interface
	return &customerService{ // --> Step 8: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
		transactor: transactor,
		custRepo:   custRepo,
		notiSvc:    notiSvc,
	}
}

// --> Step 9: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
func (s *customerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
	// ตรวจสอบเงื่อนไขของ business rules
	// Rule 1: credit ต้องมากกว่า 0
	if req.Credit <= 0 {
		return nil, ErrCreditValue
	}

	// Rule 2: ตรวจสอบ email ต้องไม่ซ้ำ
	exists, err := s.custRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		// error logging
		logger.Log.Error(err.Error())
		return nil, err
	}

	if exists {
		return nil, ErrEmailExists
	}

	// แปลง DTO → Model
	customer := model.NewCustomer(req.Email, req.Credit)

	// ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
	err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {

		// ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
		if err := s.custRepo.Create(ctx, customer); err != nil {
			// error logging
			logger.Log.Error(err.Error())
			return err
		}

		// เพิ่มส่งอีเมลต้อนรับ เข้าไปใน hook แทน การเรียกใช้งานทันที
		registerPostCommitHook(func(ctx context.Context) error {
			return s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
				"message": "Thank you for joining us! We are excited to have you as a member."})
		})

		return nil
	})

	// จัดการ error จากใน transactor
	if err != nil {
		return nil, err
	}

	// สร้าง DTO Response
	resp := dto.NewCreateCustomerResponse(customer.ID)

	return resp, nil
}
