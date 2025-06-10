package service

import (
	"context"
	"go-mma/dto"
	"go-mma/model"
	"go-mma/repository"
	"go-mma/util/errs"
	"go-mma/util/logger"
)

var (
	ErrCreditValue = errs.BusinessRuleError("credit must be greater than 0")
	ErrEmailExists = errs.ConflictError("email already exists")
)

type CustomerService struct {
	custRepo *repository.CustomerRepository
	notiSvc  *NotificationService
}

func NewCustomerService(custRepo *repository.CustomerRepository, notiSvc *NotificationService) *CustomerService {
	return &CustomerService{
		custRepo: custRepo,
		notiSvc:  notiSvc,
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

	// ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
	if err := s.custRepo.Create(ctx, customer); err != nil {
		// error logging
		logger.Log.Error(err.Error())
		return nil, err
	}

	// ส่งอีเมลต้อนรับ // <-- เพิ่มตรงนี้
	if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
		"message": "Thank you for joining us! We are excited to have you as a member.",
	}); err != nil {
		// error logging
		logger.Log.Error(err.Error())
		return nil, err
	}

	// สร้าง DTO Response
	resp := dto.NewCreateCustomerResponse(customer.ID)

	return resp, nil
}
