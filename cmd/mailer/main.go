package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/josnelihurt/mailer-go/pkg/config"

	"gopkg.in/gomail.v2"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

func main() {
	config, err := config.Read()
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	log.Printf("\nconfig ok: %v", config.String())
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		log.Printf("\nwaitting for events")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				time.Sleep(10 * time.Millisecond) //IDK it the other process writes immediatly
				if event.Op == fsnotify.Create {
					fileToSend, err := eventInfo(event.Name)
					if err != nil {
						log.Println("unkown err: ", err)
						continue
					}
					sendFile(config, fileToSend)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}

	}()

	if _, err := os.Stat(config.Inbox); os.IsNotExist(err) {
		log.Fatal("inbox folder doesn't exist", err)
	}
	if _, err := os.Stat(config.ErrBox); os.IsNotExist(err) {
		log.Fatal("err folder doesn't exist", err)
	}
	if _, err := os.Stat(config.DoneBox); os.IsNotExist(err) {
		log.Fatal("done folder doesn't exist", err)
	}

	files, err := os.ReadDir(config.Inbox)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		log.Printf("\n sending %s", file)
		fileInfo, err := file.Info()
		if err != nil {
			log.Fatal("unable to get info", err)
		}
		sendFile(config, fileInfo)
	}
	err = watcher.Add(config.Inbox)
	if err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done

}

func eventInfo(filename string) (fs.FileInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return fileInfo, nil
}

func sendFile(config config.Config, file fs.FileInfo) {
	filePath := filepath.Join(config.Inbox, file.Name())
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("\nunable to read file: %v, :%s", err, file)
	}
	msg := string(body)
	err = sendEmail(config, msg)
	if err != nil {
		log.Println("Failed to send email:", err)
	}
	newPath := filepath.Join(config.DoneBox, file.Name())

	err = os.Rename(filePath, newPath)
	if err != nil {
		log.Println("Failed to send email:", err)
	}
	log.Printf("\nnew email sent")
}

var (
	imeiToPhone = map[string]string{"355270044616142": "3044247910"}
)

func sendEmail(config config.Config, fileContent string) error {
	// Create new email message
	message := gomail.NewMessage()
	sms := parseSMSMessage(fileContent)
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
	`, sms.Message, fileContent))

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
	From     string `json:"from"`
	FromTOA  string `json:"from_toa"`
	FromSMSC string `json:"from_smsc"`
	Sent     string `json:"sent"`
	Received string `json:"received"`
	Subject  string `json:"subject"`
	Modem    string `json:"modem"`
	IMSI     string `json:"imsi"`
	IMEI     string `json:"imei"`
	Report   string `json:"report"`
	Alphabet string `json:"alphabet"`
	Length   int    `json:"length"`
	Message  string `json:"message"`
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
