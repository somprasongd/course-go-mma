package handler

import (
	"context"
	"database/sql"
	"fmt"
	"go-mma/util/idgen"
	"go-mma/util/logger"
	"go-mma/util/storage/sqldb"
	"net/mail"
	"time"

	"github.com/gofiber/fiber/v3"
)

type CustomerHandler struct {
	dbCtx sqldb.DBContext
}

func NewCustomerHandler(db sqldb.DBContext) *CustomerHandler {
	return &CustomerHandler{dbCtx: db}
}

func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
	// กำหนด payload structure (DTO: Request)
	type CreateCustomerRequest struct {
		Email  string `json:"email"`
		Credit int    `json:"credit"`
	}
	// แปลง request body -> dto
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

	// ตรวจสอบเงื่อนไขของ business rules
	// Rule 1: credit ต้องมากกว่า 0
	if req.Credit <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "credit must be greater than 0"})
	}

	// Rule 2: ตรวจสอบ email ต้องไม่ซ้ำ
	query := "SELECT 1 FROM public.customers where email = $1 LIMIT 1"
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var exists int
	if err := h.dbCtx.DB().QueryRowxContext(ctx, query, req.Email).Scan(&exists); err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error(fmt.Sprintf("error checking email: %v", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an error occurred while checking email"})
		}
	}
	if exists == 1 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already exists"})
	}

	// บันทึกลงฐานข้อมูล
	var id int64
	query = "INSERT INTO customers (id, email, credit) VALUES ($1, $2, $3) RETURNING id"
	ctxIns, cancelIns := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancelIns()
	if err := h.dbCtx.DB().QueryRowxContext(ctxIns, query, idgen.GenerateTimeRandomID(), req.Email, req.Credit).Scan(&id); err != nil {
		logger.Log.Error(fmt.Sprintf("error insert customer: %v", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an error occurred while inserting customer"})
	}

	// กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป (DTO: Response)
	type CreateCustomerResponse struct {
		ID int64 `json:"id"`
	}
	resp := &CreateCustomerResponse{ID: id}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}
