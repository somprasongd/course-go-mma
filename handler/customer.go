package handler

import (
	"fmt"
	"go-mma/util/logger"
	"net/mail"

	"github.com/gofiber/fiber/v3"
)

type CustomerHandler struct {
}

func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{}
}

func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
	// กำหนด payload structure
	type CreateCustomerRequest struct {
		Email  string `json:"email"`
		Credit int    `json:"credit"`
	}
	// แปลง request body -> struct
	var req CreateCustomerRequest
	if err := c.Bind().Body(&req); err != nil {
		// แปลงไม่ได้ให้ส่ง error 400
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Log.Info(fmt.Sprintf("Received customer: %v", req))

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is invalid"})
	}

	// ตรวจสอบ business rules (e.g., credit must be greater than 0)
	if req.Credit <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "credit must be greater than 0"})
	}

	// TODO: บันทึกลงฐานข้อมูล
	var id int // id ในฐานข้อมูล

	// กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป
	type CreateCustomerResponse struct {
		ID int `json:"id"`
	}
	resp := &CreateCustomerResponse{ID: id}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}
