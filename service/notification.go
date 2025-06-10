package service

import (
	"fmt"
	"go-mma/util/logger"
)

type NotificationService struct {
}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) SendEmail(to string, subject string, payload map[string]any) error {
	// implement email sending logic here
	logger.Log.Info(fmt.Sprintf("Sending email to %s with subject: %s and payload: %v", to, subject, payload))
	return nil
}
