package mailer

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/josnelihurt/mailer-go/pkg/config"
)

func SendFile(cfg config.Config, file fs.FileInfo) {
	filePath := filepath.Join(cfg.Inbox, file.Name())
	body, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("\nunable to read file: %v, :%s", err, file)
	}
	msg := string(body)

	sms := parseSMSMessage(msg)
	sms.SMSToolsFile = msg

	err = SendEmail(cfg, sms)
	if err != nil {
		log.Println("Failed to send email:", err)
	}

	//This bad practice, I know it but I'm lazy to change it now
	client := New(cfg)
	err = client.SendToServer("incoming", msg)
	if err != nil {
		log.Println("Failed to enqueue on server queue:", err)
	}

	newPath := filepath.Join(cfg.DoneBox, file.Name())

	err = os.Rename(filePath, newPath)
	if err != nil {
		log.Println("Failed to move file:", err)
	}
	log.Printf("new email sent")
}
