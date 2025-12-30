package mailer

import (
	"fmt"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"gopkg.in/gomail.v2"
)

const (
	smtpAuthAddress = "smtp.gmail.com"
)

func SendEmail(cfg config.Config, sms SMSMessage) error {
	message := gomail.NewMessage()
	message.SetHeader("From", cfg.Email)
	message.SetHeader("To", cfg.RecipientEmail...)
	sentTo := cfg.ImeiToPhone[sms.IMEI]
	message.SetHeader("Subject", fmt.Sprintf("[SMS DATE TO] %s %s", sms.Sent, sentTo))
	message.SetBody("text/plain", fmt.Sprintf(`
		%s




----------------------------
FULL MESSAGE FROM FILE:
%s
----------------------------
	`, sms.Message, sms.SMSToolsFile))

	dialer := gomail.NewDialer(smtpAuthAddress, 587, cfg.Email, cfg.Password)

	err := dialer.DialAndSend(message)
	if err != nil {
		return err
	}

	return nil
}
