package mailer

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

func parseSMSMessage(messageText string) SMSMessage {
	var message SMSMessage

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
