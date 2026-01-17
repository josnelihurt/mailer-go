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

	// Get phone number from IMEI mapping
	sentTo := cfg.ImeiToPhone[sms.IMEI]
	if sentTo == "" {
		sentTo = sms.From
	}

	subject := fmt.Sprintf("[SMS] %s -> %s", sms.From, sentTo)
	if sms.Sent != "" {
		subject = fmt.Sprintf("[SMS %s] %s -> %s", sms.Sent, sms.From, sentTo)
	}

	message.SetHeader("Subject", subject)

	body := fmt.Sprintf(`From: %s
To: %s
Received: %s

Message:
%s

---
Device: %s
IMEI: %s
Length: %d
Alphabet: %s
`,
		sms.From,
		sentTo,
		sms.Received,
		sms.Message,
		sms.ModemDevice,
		sms.IMEI,
		sms.Length,
		sms.Alphabet,
	)

	message.SetBody("text/plain", body)

	dialer := gomail.NewDialer(smtpAuthAddress, 587, cfg.Email, cfg.Password)

	err := dialer.DialAndSend(message)
	if err != nil {
		return err
	}

	return nil
}
