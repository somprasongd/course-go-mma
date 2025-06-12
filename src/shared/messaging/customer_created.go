package messaging

import (
	"go-mma/shared/common/eventbus"
	"go-mma/shared/common/idgen"
	"time"
)

const (
	CustomerCreatedIntegrationEventName eventbus.EventName = "CustomerCreated"
)

type CustomerCreatedIntegrationEvent struct {
	eventbus.BaseEvent
	CustomerID int64  `json:"customer_id"`
	Email      string `json:"email"`
}

func NewCustomerCreatedIntegrationEvent(customerID int64, email string) *CustomerCreatedIntegrationEvent {
	return &CustomerCreatedIntegrationEvent{
		BaseEvent: eventbus.BaseEvent{
			ID:   idgen.GenerateUUIDLikeID(),
			Name: CustomerCreatedIntegrationEventName,
			At:   time.Now(),
		},
		CustomerID: customerID,
		Email:      email,
	}
}
