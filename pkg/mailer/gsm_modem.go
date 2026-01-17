package mailer

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/tarm/serial"
)

const (
	telegramPollInterval = 5 * time.Second
	readTimeout        = 10 * time.Second
)

type GSMModem struct {
	device string
	port   *serial.Port
	cfg    config.Config
	imei   string
}

func NewGSMModem(cfg config.Config) (*GSMModem, error) {
	serialCfg := &serial.Config{
		Name:        cfg.ModemDevice,
		Baud:        cfg.ModemBaud,
		ReadTimeout: readTimeout,
	}

	port, err := serial.OpenPort(serialCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open modem: %w", err)
	}

	return &GSMModem{
		device: cfg.ModemDevice,
		port:   port,
		cfg:    cfg,
	}, nil
}

func (m *GSMModem) Initialize() error {
	log.Println("Initializing GSM modem...")

	// Send commands just like my Python script
	type cmd struct {
		cmd  string
		desc string
	}
	commands := []cmd{
		{"AT", "Check modem responsive"},
		{"ATE0", "Disable echo"},
		{"AT+CMEE=2", "Verbose errors"},
		{"AT+CMGF=1", "Text mode"},
		{"AT+CSCS=\"UCS2\"", "UCS2 charset"},
		{"AT+CNMI=2,1,0,0,0", "New message indication"},
	}

	for _, cmd := range commands {
		resp, err := m.sendCommand(cmd.cmd)
		if err != nil {
			return fmt.Errorf("%s failed: %w", cmd.desc, err)
		}
		if !strings.Contains(resp, "OK") {
			return fmt.Errorf("%s returned error: %s", cmd.desc, resp)
		}
		log.Printf("✓ %s", cmd.desc)
	}

	// Get IMEI
	resp, err := m.sendCommand("AT+CGSN")
	if err != nil {
		return fmt.Errorf("failed to get IMEI: %w", err)
	}

	// Parse IMEI from response (first 15 digit number)
	re := regexp.MustCompile(`\d{15}`)
	matches := re.FindStringSubmatch(resp)
	if matches != nil {
		m.imei = matches[0]
		log.Printf("Modem IMEI: %s", m.imei)
	}

	return nil
}

func (m *GSMModem) sendCommand(cmd string) (string, error) {
	// Clear buffer
	if err := m.port.Flush(); err != nil {
		return "", err
	}

	// Send command
	if _, err := m.port.Write([]byte(cmd + "\r")); err != nil {
		return "", err
	}

	// Read response with timeout
	buf := make([]byte, 4096)
	response := ""
	deadline := time.Now().Add(readTimeout)

	for time.Now().Before(deadline) {
		n, err := m.port.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Timeout is OK
			if strings.Contains(err.Error(), "timeout") {
				break
			}
			return "", err
		}

		if n > 0 {
			response += string(buf[:n])

			// Check if we received OK or ERROR
			if strings.Contains(response, "OK") || strings.Contains(response, "ERROR") {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	return response, nil
}

func (m *GSMModem) ListMessages() ([]SMSMessage, error) {
	log.Printf("Sending AT+CMGL command...")
	resp, err := m.sendCommand("AT+CMGL=\"ALL\"")
	if err != nil {
		log.Printf("AT+CMGL command error: %v", err)
		return nil, err
	}

	log.Printf("AT+CMGL response: %s", resp)
	if !strings.Contains(resp, "OK") {
		log.Printf("AT+CMGL failed, response: %s", resp)
		return nil, fmt.Errorf("CMGL failed: %s", resp)
	}

	messages := m.ParseCMGLResponse(resp)
	log.Printf("Parsed %d messages from CMGL response", len(messages))
	return messages, nil
}

func (m *GSMModem) ParseCMGLResponse(data string) []SMSMessage {
	lines := strings.Split(data, "\n")
	var messages []SMSMessage

	i := 0
	for i < len(lines) {
		line := lines[i]

	// Look for CMGL line
		if strings.Contains(line, "+CMGL:") {
			// Format: +CMGL: 0,"REC READ","002B0031...",,"26/01/17,00:59:09-20"
			// Use simple string parsing instead of regex
			parts := strings.Split(line, ",")
			if len(parts) >= 4 {
				// Extract index from +CMGL: 0,
				indexStr := strings.TrimSpace(strings.Split(parts[0], ":")[1])
				index, _ := strconv.Atoi(indexStr)

				// parts[1] = "REC READ"
				// parts[2] = "002B0031..." (phone number in UCS2)
				phoneNumber := DecodeUCS2(strings.Trim(parts[2], `"`))

				// parts[3] is empty, parts[4] is timestamp
				timestamp := ""
				if len(parts) >= 5 {
					timestamp = strings.Trim(parts[4], `"`)
				}

				log.Printf("Parsing CMGL index=%d, phone=%s, timestamp=%s", index, phoneNumber, timestamp)

				// Message content is on the next line(s)
				messageContent := ""
				i++
				for i < len(lines) && !strings.Contains(lines[i], "OK") && !strings.Contains(lines[i], "+CMGL:") {
					if strings.TrimSpace(lines[i]) != "" {
						messageContent = lines[i]
						break
					}
					i++
				}

				decodedMessage := DecodeUCS2(messageContent)
				log.Printf("Decoded message: %s", decodedMessage)

				sms := SMSMessage{
					From:        phoneNumber,
					Message:     decodedMessage,
					Sent:        timestamp,
					Received:    time.Now().Format("2006-01-02 15:04:05"),
					ModemDevice: m.device,
					Modem:       m.device,
					Index:       index,
					Alphabet:    "UCS2",
					Length:      len(decodedMessage),
					IMEI:        m.imei,
				}
				messages = append(messages, sms)
			} else {
				log.Printf("Failed to parse CMGL line (not enough parts): %s", line)
			}
		} else {
			i++
		}
	}

	return messages
}

func DecodeUCS2(hex_string string) string {
	// Remove any whitespace
	hex_string = strings.ReplaceAll(hex_string, " ", "")
	hex_string = strings.TrimSpace(hex_string)

	// If odd length, remove last character (might be extra byte from modem)
	if len(hex_string)%2 != 0 {
		hex_string = hex_string[:len(hex_string)-1]
	}

	// Convert hex to bytes
	bytes, err := hex.DecodeString(hex_string)
	if err != nil {
		log.Printf("UCS2 hex decode error: %v", err)
		return hex_string
	}

	// Decode UCS-2 Big Endian to runes
	runes := make([]rune, 0)
	for i := 0; i < len(bytes); i += 2 {
		if i+1 < len(bytes) {
			// UCS-2 Big Endian: high byte first
			r := rune(bytes[i])<<8 | rune(bytes[i+1])
			runes = append(runes, r)
		}
	}

	return string(runes)
}

func (m *GSMModem) Start(messageHandler func(SMSMessage)) {
	log.Printf("Starting SMS polling every %s", telegramPollInterval)

	// Process existing messages first
	messages, err := m.ListMessages()
	if err != nil {
		log.Printf("Failed to list messages: %v", err)
	} else {
		log.Printf("Processing %d existing messages", len(messages))
		for _, msg := range messages {
			messageHandler(msg)
			if m.cfg.DeleteAfterRead {
				m.DeleteMessage(msg.Index)
			}
		}
	}

	// Start polling loop
	ticker := time.NewTicker(telegramPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Polling for messages...")
			messages, err := m.ListMessages()
			if err != nil {
				log.Printf("Failed to list messages: %v", err)
				continue
			}

			log.Printf("Poll complete: %d messages found", len(messages))
			if len(messages) > 0 {
				log.Printf("Found %d messages", len(messages))
				for _, msg := range messages {
					messageHandler(msg)
					if m.cfg.DeleteAfterRead {
						m.DeleteMessage(msg.Index)
					}
				}
			}
		}
	}
}

func (m *GSMModem) DeleteMessage(index int) error {
	resp, err := m.sendCommand(fmt.Sprintf("AT+CMGD=%d,4", index))
	if err != nil {
		return err
	}

	if !strings.Contains(resp, "OK") {
		return fmt.Errorf("delete failed: %s", resp)
	}

	log.Printf("Deleted message %d", index)
	return nil
}

func (m *GSMModem) GetIMEI() string {
	return m.imei
}
