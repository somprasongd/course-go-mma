package service

import (
	"context"
	"go-mma/modules/order/dto"
	"go-mma/modules/order/model"
	"go-mma/modules/order/repository"
	"go-mma/util/errs"
	"go-mma/util/logger"
	"go-mma/util/storage/sqldb/transactor"

	custRepository "go-mma/modules/customer/repository"
	notiService "go-mma/modules/notification/service"
)

var (
	ErrTotalOrderValue = errs.BusinessRuleError("order_total must be greater than 0")
	ErrNoCustomerID    = errs.ResourceNotFoundError("the customer with given id was not found")
	ErrNoOrderID       = errs.ResourceNotFoundError("the order with given id was not found")
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
	CancelOrder(ctx context.Context, id int64) error
}

type orderService struct {
	transactor transactor.Transactor
	custRepo   custRepository.CustomerRepository
	orderRepo  repository.OrderRepository
	notiSvc    notiService.NotificationService
}

func NewOrderService(
	transactor transactor.Transactor,
	custRepo custRepository.CustomerRepository,
	orderRepo repository.OrderRepository,
	notiSvc notiService.NotificationService,
) OrderService {
	return &orderService{
		transactor: transactor,
		custRepo:   custRepo,
		orderRepo:  orderRepo,
		notiSvc:    notiSvc,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	// Business Logic Rule: ตรวจสอบ ยอดรวมต้องมากกว่า 0
	if req.OrderTotal <= 0 {
		return nil, ErrTotalOrderValue
	}

	// Business Logic Rule: ตรวจสอบ customer id
	customer, err := s.custRepo.FindByID(ctx, req.CustomerID)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, err
	}

	if customer == nil {
		return nil, ErrNoCustomerID
	}

	// Business Logic Rule: ตัดยอด credit ถ้าไม่พอให้ error
	if err := customer.ReserveCredit(req.OrderTotal); err != nil {
		return nil, err
	}

	// ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
	var order *model.Order
	err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
		// ตัดยอด credit ในตาราง customer
		if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		// สร้าง order ใหม่ DTO -> Model
		order = model.NewOrder(req.CustomerID, req.OrderTotal)
		// บันทึกลงฐานข้อมูล
		err = s.orderRepo.Create(ctx, order)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		// ส่งอีเมลยืนยัน
		registerPostCommitHook(func(ctx context.Context) error {
			return s.notiSvc.SendEmail(customer.Email, "Order Created", map[string]any{
				"order_id": order.ID,
				"total":    order.OrderTotal,
			})
		})

		return nil
	})

	// จัดการ error จากใน transactor
	if err != nil {
		return nil, err
	}

	// สร้าง DTO Response
	resp := dto.NewCreateOrderResponse(order.ID)
	return resp, nil
}

func (s *orderService) CancelOrder(ctx context.Context, id int64) error {
	// Business Logic Rule: ตรวจสอบ order id
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	if order == nil {
		return ErrNoOrderID
	}

	// ยกเลิก order
	if err := s.orderRepo.Cancel(ctx, order.ID); err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	// Business Logic Rule: ตรวจสอบ customer id
	customer, err := s.custRepo.FindByID(ctx, order.CustomerID)
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	if customer == nil {
		return ErrNoCustomerID
	}

	// Business Logic: คืนยอด credit
	customer.ReleaseCredit(order.OrderTotal)

	// บันทึกการคืนยอด credit
	if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	return nil
}
