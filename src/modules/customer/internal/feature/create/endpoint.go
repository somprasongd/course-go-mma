package create

import (
	"fmt"
	"go-mma/shared/common/errs"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/mediator"

	"github.com/gofiber/fiber/v3"
)

func NewEndpoint(router fiber.Router, path string) {
	router.Post(path, createCustomerHTTPHandler)
}

func createCustomerHTTPHandler(c fiber.Ctx) error {
	// แปลง request body -> dto
	var req CreateCustomerRequest
	if err := c.Bind().Body(&req); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	logger.Log.Info(fmt.Sprintf("Received customer: %v", req))

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if err := req.Validate(); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	// *** ส่งไปที่ Command Handler แทน Service ***
	resp, err := mediator.Send[*CreateCustomerCommand, *CreateCustomerCommandResult](
		c.Context(),
		&CreateCustomerCommand{CreateCustomerRequest: req},
	)

	// จัดการ error จาก Service Layer หากเกิดขึ้น
	if err != nil {
		// จัดการ error response ที่ middleware
		return err
	}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}
