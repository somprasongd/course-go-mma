package cancel

import (
	"fmt"
	"go-mma/shared/common/errs"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/mediator"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func NewEndpoint(router fiber.Router, path string) {
	router.Delete(path, cancelOrderHTTPHandler)
}

// CancelOrder godoc
// @Summary		Cancel Order
// @Description	Cancel Order By Order ID
// @Tags			Order
// @Produce		json
// @Param			orderID	path	string	true	"order id"
// @Failure		401
// @Failure		404
// @Failure		500
// @Success		204
// @Router			/orders/{orderID} [delete]
func cancelOrderHTTPHandler(c fiber.Ctx) error {
	// ตรวจสอบรูปแบบ orderID
	orderID, err := strconv.Atoi(c.Params("orderID"))
	if err != nil {
		// จัดการ error response ที่ middleware
		return errs.InputValidationError("invalid order id")
	}

	logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))

	// ส่งไปที่ Command Handler
	_, err = mediator.Send[*CancelOrderCommand, *mediator.NoResponse](
		c.Context(),
		&CancelOrderCommand{ID: int64(orderID)},
	)

	// จัดการ error จาก feature หากเกิดขึ้น
	if err != nil {
		// จัดการ error response ที่ middleware
		return err
	}

	// ตอบกลับด้วย status code 204 (no content)
	return c.SendStatus(fiber.StatusNoContent)
}
