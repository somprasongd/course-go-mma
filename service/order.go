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
	ErrTotalOrderValue = errs.BusinessRuleError("order_total must be greater than 0")
	ErrNoCustomerID    = errs.ResourceNotFoundError("the customer with given id was not found")
	ErrNoOrderID       = errs.ResourceNotFoundError("the order with given id was not found")
)

type OrderService struct {
	custRepo  *repository.CustomerRepository
	orderRepo *repository.OrderRepository
	notiSvc   *NotificationService
}

func NewOrderService(custRepo *repository.CustomerRepository, orderRepo *repository.OrderRepository, notiSvc *NotificationService) *OrderService {
	return &OrderService{
		custRepo:  custRepo,
		orderRepo: orderRepo,
		notiSvc:   notiSvc,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
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

	// ตัดยอด credit ในตาราง customer
	if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
		logger.Log.Error(err.Error())
		return nil, err
	}

	// สร้าง order ใหม่ DTO -> Model
	order := model.NewOrder(req.CustomerID, req.OrderTotal)
	// บันทึกลงฐานข้อมูล
	err = s.orderRepo.Create(ctx, order)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, err
	}

	// ส่งอีเมลยืนยัน
	err = s.notiSvc.SendEmail(customer.Email, "Order Created", map[string]any{
		"order_id": order.ID,
		"total":    order.OrderTotal,
	})
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, err
	}

	// สร้าง DTO Response
	resp := dto.NewCreateOrderResponse(order.ID)
	return resp, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, id int64) error {
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
