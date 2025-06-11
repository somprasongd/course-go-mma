package service

import (
	"context"
	"go-mma/modules/customer/dto"
	"go-mma/modules/customer/model"
	"go-mma/modules/customer/repository"
	"go-mma/util/errs"
	"go-mma/util/logger"
	"go-mma/util/storage/sqldb/transactor"

	notiService "go-mma/modules/notification/service"
)

var (
	ErrCreditValue = errs.BusinessRuleError("credit must be greater than 0")
	ErrEmailExists = errs.ConflictError("email already exists")
)

type CustomerService interface {
	CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error)
}

type customerService struct {
	transactor transactor.Transactor
	custRepo   repository.CustomerRepository
	notiSvc    notiService.NotificationService
}

func NewCustomerService(
	transactor transactor.Transactor,
	custRepo repository.CustomerRepository,
	notiSvc notiService.NotificationService,
) CustomerService {
	return &customerService{
		transactor: transactor,
		custRepo:   custRepo,
		notiSvc:    notiSvc,
	}
}

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
