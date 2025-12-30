package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	serverURL  string
	apiKey     string
	httpClient *http.Client
}

type SMSEnqueueRequest struct {
	SMSMessage SMSMessage `json:"sms_message"`
	FolderName string     `json:"folder_name"`
}

func New() *Client {
	serverURL := os.Getenv("MAILER_SERVER_URL")
	if serverURL == "" {
		serverURL = "https://mailer-go.vultur.josnelihurt.me"
	}

	apiKey := os.Getenv("MAILER_API_KEY")
	if apiKey == "" {
		log.Fatal("MAILER_API_KEY environment variable is required in client mode")
	}

	return &Client{
		serverURL: serverURL,
		apiKey:    apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendToServer(folderName string, content string) error {
	// Parse SMS message from content
	sms := parseSMSMessage(content)
	sms.SMSToolsFile = content // Store raw content in case parse fails

	req := SMSEnqueueRequest{
		SMSMessage: sms,
		FolderName: folderName,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.serverURL+"/v1/sms/enqueue", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("successfully enqueued on server queue")

	return nil
}
