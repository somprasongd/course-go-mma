package handler

import (
	"fmt"
	"go-mma/dto"
	"go-mma/service"
	"go-mma/util/errs"
	"go-mma/util/logger"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type OrderHandler struct {
	orderSvc *service.OrderService
}

func NewOrderHandler(orderSvc *service.OrderService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc}
}

func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	// แปลง request body -> struct
	var req dto.CreateOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	logger.Log.Info(fmt.Sprintf("Received Order: %v", req))

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if err := req.Validate(); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	// ส่งไปที่ Service Layer
	resp, err := h.orderSvc.CreateOrder(c.Context(), &req)

	// จัดการ error จาก Service Layer หากเกิดขึ้น
	if err != nil {
		// จัดการ error response ที่ middleware
		return err
	}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *OrderHandler) CancelOrder(c fiber.Ctx) error {
	// ตรวจสอบรูปแบบ orderID
	orderID, err := strconv.Atoi(c.Params("orderID"))
	if err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError("invalid order id")
	}

	logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))

	// ส่งไปที่ Service Layer
	err = h.orderSvc.CancelOrder(c.Context(), int64(orderID))

	// จัดการ error จาก Service Layer หากเกิดขึ้น
	if err != nil {
		// จัดการ error response ที่ middleware
		return err
	}

	// ตอบกลับด้วย status code 204 (no content)
	return c.SendStatus(fiber.StatusNoContent)
}
