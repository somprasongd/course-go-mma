package service

import (
	"fmt"
	"go-mma/shared/common/logger"
)

// --> Step 1: สร้าง interface
type NotificationService interface {
	SendEmail(to string, subject string, payload map[string]any) error
}

// --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
type notificationService struct {
}

// --> Step 3: return เป็น interface
func NewNotificationService() NotificationService {
	return &notificationService{} // --> Step 4: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
}

// --> Step 5: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
func (s *notificationService) SendEmail(to string, subject string, payload map[string]any) error {
	// implement email sending logic here
	logger.Log.Info(fmt.Sprintf("Sending email to %s with subject: %s and payload: %v", to, subject, payload))
	return nil
}
