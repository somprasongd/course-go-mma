package handler

import (
	"fmt"
	"go-mma/dto"
	"go-mma/service"
	"go-mma/util/logger"

	"github.com/gofiber/fiber/v3"
)

type CustomerHandler struct {
	custService *service.CustomerService
}

func NewCustomerHandler(custService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		custService: custService,
	}
}

func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
	// แปลง request body -> dto
	var req dto.CreateCustomerRequest
	if err := c.Bind().Body(&req); err != nil {
		// แปลงไม่ได้ให้ส่ง error 400
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Log.Info(fmt.Sprintf("Received customer: %v", req))

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// ส่งไปที่ Service Layer
	resp, err := h.custService.CreateCustomer(c.Context(), &req)

	// จัดการ error จาก Service Layer หากเกิดขึ้น
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}
