package create

import (
	"go-mma/shared/common/errs"
	"go-mma/shared/common/mediator"

	"github.com/gofiber/fiber/v3"
)

func NewEndpoint(router fiber.Router, path string) {
	router.Post(path, createOrderHTTPHandler)
}

// CreateOrder godoc
// @Summary		Create Order
// @Description	Create Order
// @Tags			Order
// @Produce		json
// @Param			order	body	CreateOrderRequest	true	"Create Data"
// @Failure		401
// @Failure		500
// @Success		201	{object}	CreateOrderResponse
// @Router			/orders [post]
func createOrderHTTPHandler(c fiber.Ctx) error {
	// แปลง request body -> struct
	var req CreateOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if err := req.Validate(); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	// ส่งไปที่ Command Handler
	resp, err := mediator.Send[*CreateOrderCommand, *CreateOrderCommandResult](
		c.Context(),
		&CreateOrderCommand{CreateOrderRequest: req},
	)

	// จัดการ error จาก feature หากเกิดขึ้น
	if err != nil {
		// จัดการ error response ที่ middleware
		return err
	}

	// ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
	return c.Status(fiber.StatusCreated).JSON(resp)
}
