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

type CustomerService struct {
	transactor transactor.Transactor
	custRepo   *repository.CustomerRepository
	notiSvc    *NotificationService
}

func NewCustomerService(transactor transactor.Transactor, custRepo *repository.CustomerRepository, notiSvc *NotificationService) *CustomerService {
	return &CustomerService{
		transactor: transactor,
		custRepo:   custRepo,
		notiSvc:    notiSvc,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
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

	// ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction // <-- ตรงนี้
	err = s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {

		// ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
		if err := s.custRepo.Create(ctx, customer); err != nil {
			// error logging
			logger.Log.Error(err.Error())
			return err
		}

		// ส่งอีเมลต้อนรับ // <-- เพิ่มตรงนี้
		if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
			"message": "Thank you for joining us! We are excited to have you as a member.",
		}); err != nil {
			// error logging
			logger.Log.Error(err.Error())
			return err
		}
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
