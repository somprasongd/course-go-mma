package create

import (
	"fmt"
	"go-mma/shared/common/errs"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/mediator"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel/trace"
)

func NewEndpoint(router fiber.Router, path string) {
	router.Post(path, createCustomerHTTPHandler)
}

// CreateCustomer godoc
// @Summary		Create Customer
// @Description	Create Customer
// @Tags			Customer
// @Produce		json
// @Param			customer	body	CreateCustomerRequest	true	"Create Data"
// @Failure		401
// @Failure		500
// @Success		201	{object}	CreateCustomerResponse
// @Router			/customers [post]
func createCustomerHTTPHandler(c fiber.Ctx) error {
	ctx := c.Context()
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("http_handler")
	ctx, span := tracer.Start(ctx, "Endpoint:CreateCustomer")
	defer span.End()

	// แปลง request body -> dto
	var req CreateCustomerRequest
	if err := c.Bind().Body(&req); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	logger.FromContext(ctx).Info(fmt.Sprintf("Received customer: %v", req))

	// ตรวจสอบ input fields (e.g., value, format, etc.)
	if err := req.Validate(); err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError(err.Error())
	}

	// *** ส่งไปที่ Command Handler แทน Service ***
	resp, err := mediator.Send[*CreateCustomerCommand, *CreateCustomerCommandResult](
		ctx,
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
