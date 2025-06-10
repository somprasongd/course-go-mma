package handler

import (
	"fmt"
	"go-mma/util/logger"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type OrderHandler struct {
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	// กำหนด payload structure
	type CreateOrderRequest struct {
		CustomerID string `json:"customer_id"`
		OrderTotal int    `json:"order_total"`
	}
	// แปลง request body -> struct
	var req CreateOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		// แปลงไม่ได้ให้ส่ง error 400
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Log.Info(fmt.Sprintf("Received Order: %v", req))

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if req.CustomerID == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "customer_id is required"})
	}

	// ตรวจสอบ business rules (e.g., order_total must be greater than 0)
	if req.OrderTotal <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "order_total must be greater than 0"})
	}

	// TODO: ตรวจสอบว่ามี customer id อยู่ในฐานข้อมูล หรือไม่
	// customer := getCustomer(order.CustomerID)
	// if customer == nil {
	// 	return return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "the customer with given id was not found"})
	// }

	// TODO: ตรวจสอบ credit เพียงพอ หรือไม่
	// if credit < payload.OrderTotal {
	// 	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "insufficient credit"})
	// }

	// TODO: หักยอด credit ของ customer

	// TODO: update customer's credit balance ในฐานข้อมูล

	// TODO: บันทึกรายการออเดอร์ใหม่ลงในฐานข้อมูล
	var id int

	// กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป
	type CreateOrderResponse struct {
		ID int `json:"id"`
	}
	resp := &CreateOrderResponse{ID: id}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *OrderHandler) CancelOrder(c fiber.Ctx) error {
	// ตรวจสอบรูปแบบ orderID
	orderID, err := strconv.Atoi(c.Params("orderID"))
	if err != nil {
		// ถ้าไม่ถูกต้อง error 400
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}

	logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))

	// TODO: ตรวจสอบ orderID ในฐานข้อมูล
	// order := getOrder(orderID)
	// if order == nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "the order with given id was not found"})
	// }

	// TODO: ค้นหา customer จาก customerID ของ order
	// customer := getCustomer(order.CustomerID)
	// if customer == nil {
	// 	return return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "the customer with given id was not found"})
	// }

	// TODO: คืนยอด credit กลับให้ customer
	// creditLimit += CreateOrderRequest.OrderTotal

	// TODO: update customer's credit balance ในฐานข้อมูล

	// TODO: update สถานะของ order ในฐานข้อมูลให้เป็น cancel

	// ตอบกลับด้วย status code 204 (no content)
	return c.SendStatus(fiber.StatusNoContent)
}
