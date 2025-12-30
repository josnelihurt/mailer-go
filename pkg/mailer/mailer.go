package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/redis/go-redis/v9"

	"gopkg.in/gomail.v2"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

var redisClient *redis.Client

func InitRedisClient(cfg config.Config) {
	if !cfg.RedisEnabled {
		return
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed (non-fatal): %v", err)
		redisClient = nil
	} else {
		log.Printf("Redis connected successfully at %s:%s", cfg.RedisHost, cfg.RedisPort)
	}
	status := redisClient.Set(ctx, "mailer-go:started", time.Now().Format(time.RFC3339), 0)
	if status.Err() != nil {
		log.Printf("Failed to set started timestamp in Redis: %v", status.Err())
	} else {
		log.Printf("Started timestamp set in Redis: %s", status.Val())
	}
}

func PushToRedis(cfg config.Config, folderName string, sms SMSMessage) {
	if !cfg.RedisEnabled || redisClient == nil {
		log.Printf("Redis is not enabled or client is not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Marshal to JSON
	jsonData, err := json.Marshal(sms)
	if err != nil {
		log.Printf("Failed to marshal SMS message to JSON (non-fatal): %v", err)
		return
	}

	channelName := fmt.Sprintf("sms:%s", folderName)
	err = redisClient.Publish(ctx, channelName, string(jsonData)).Err()
	if err != nil {
		log.Printf("Failed to publish to Redis channel %s (non-fatal): %v", channelName, err)
	} else {
		log.Printf("Successfully published to Redis channel: %s", channelName)
	}
}

func SendFile(config config.Config, file fs.FileInfo) {
	filePath := filepath.Join(config.Inbox, file.Name())
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("\nunable to read file: %v, :%s", err, file)
	}
	msg := string(body)

	// Parse SMS message
	sms := parseSMSMessage(msg)
	sms.SMSToolsFile = msg // Store raw content

	// Send email first (primary persistent method)
	err = SendEmail(config, sms)
	if err != nil {
		log.Println("Failed to send email:", err)
	}

	client := New()
	err = client.SendToServer("incoming", msg)
	if err != nil {
		log.Println("Failed to enqueue on server queue:", err)
	}

	newPath := filepath.Join(config.DoneBox, file.Name())

	err = os.Rename(filePath, newPath)
	if err != nil {
		log.Println("Failed to move file:", err)
	}
	log.Printf("new email sent")
}

var (
	imeiToPhone = map[string]string{
		"355270044616142": "3044247910",
		"355270044531622": "3005335930",
	}
)

func SendEmail(config config.Config, sms SMSMessage) error {
	// Create new email message
	message := gomail.NewMessage()
	message.SetHeader("From", config.Email)
	message.SetHeader("To", config.RecipientEmail...)
	sentTo := imeiToPhone[sms.IMEI]
	message.SetHeader("Subject", fmt.Sprintf("[SMS DATE TO] %s %s", sms.Sent, sentTo))
	message.SetBody("text/plain", fmt.Sprintf(`
		%s




----------------------------
FULL MESSAGE FROM FILE:
%s
----------------------------
	`, sms.Message, sms.SMSToolsFile))

	// Create SMTP dialer
	dialer := gomail.NewDialer(smtpAuthAddress, 587, config.Email, config.Password)

	// Send email
	err := dialer.DialAndSend(message)
	if err != nil {
		return err
	}

	return nil
}

type SMSMessage struct {
	From         string `json:"from"`
	FromTOA      string `json:"from_toa"`
	FromSMSC     string `json:"from_smsc"`
	Sent         string `json:"sent"`
	Received     string `json:"received"`
	Subject      string `json:"subject"`
	Modem        string `json:"modem"`
	IMSI         string `json:"imsi"`
	IMEI         string `json:"imei"`
	Report       string `json:"report"`
	Alphabet     string `json:"alphabet"`
	Length       int    `json:"length"`
	Message      string `json:"message"`
	SMSToolsFile string `json:"smstools_file"` // Raw file content in case parse fails
}

func parseSMSMessage(messageText string) SMSMessage {
	var message SMSMessage

	// Define regular expressions to match and capture the fields
	reFrom := regexp.MustCompile(`From:\s+(.+)`)
	reFromTOA := regexp.MustCompile(`From_TOA:\s+(.+)`)
	reFromSMSC := regexp.MustCompile(`From_SMSC:\s+(.+)`)
	reSent := regexp.MustCompile(`Sent:\s+(.+)`)
	reReceived := regexp.MustCompile(`Received:\s+(.+)`)
	reSubject := regexp.MustCompile(`Subject:\s+(.+)`)
	reModem := regexp.MustCompile(`Modem:\s+(.+)`)
	reIMSI := regexp.MustCompile(`IMSI:\s+(.+)`)
	reIMEI := regexp.MustCompile(`IMEI:\s+(.+)`)
	reReport := regexp.MustCompile(`Report:\s+(.+)`)
	reAlphabet := regexp.MustCompile(`Alphabet:\s+(.+)`)
	reLength := regexp.MustCompile(`Length:\s+(\d+)`)

	lines := strings.Split(messageText, "\n")

	for _, line := range lines {
		if matches := reFrom.FindStringSubmatch(line); matches != nil {
			message.From = strings.TrimSpace(matches[1])
		} else if matches := reFromTOA.FindStringSubmatch(line); matches != nil {
			message.FromTOA = strings.TrimSpace(matches[1])
		} else if matches := reFromSMSC.FindStringSubmatch(line); matches != nil {
			message.FromSMSC = strings.TrimSpace(matches[1])
		} else if matches := reSent.FindStringSubmatch(line); matches != nil {
			message.Sent = strings.TrimSpace(matches[1])
		} else if matches := reReceived.FindStringSubmatch(line); matches != nil {
			message.Received = strings.TrimSpace(matches[1])
		} else if matches := reSubject.FindStringSubmatch(line); matches != nil {
			message.Subject = strings.TrimSpace(matches[1])
		} else if matches := reModem.FindStringSubmatch(line); matches != nil {
			message.Modem = strings.TrimSpace(matches[1])
		} else if matches := reIMSI.FindStringSubmatch(line); matches != nil {
			message.IMSI = strings.TrimSpace(matches[1])
		} else if matches := reIMEI.FindStringSubmatch(line); matches != nil {
			message.IMEI = strings.TrimSpace(matches[1])
		} else if matches := reReport.FindStringSubmatch(line); matches != nil {
			message.Report = strings.TrimSpace(matches[1])
		} else if matches := reAlphabet.FindStringSubmatch(line); matches != nil {
			message.Alphabet = strings.TrimSpace(matches[1])
		} else if matches := reLength.FindStringSubmatch(line); matches != nil {
			message.Length = parseInt(matches[1])
		} else if line != "" {
			// Assume the remaining non-empty lines are the message
			message.Message += line + "\n"
		}
	}

	message.Message = strings.TrimSpace(message.Message)

	return message
}

func parseInt(s string) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		log.Println("Failed to parse integer:", err)
	}
	return i
}
