package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/josnelihurt/mailer-go/pkg/config"
)

type Client struct {
	serverURL  string
	apiKey     string
	httpClient *http.Client
}

func New(cfg config.Config) *Client {
	apiKey := cfg.APIKey
	if apiKey == "" {
		log.Fatal("api_key configuration is required in client mode")
	}

	return &Client{
		serverURL: cfg.ServerURL,
		apiKey:    apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendToServer(folderName string, content string) error {
	sms := parseSMSMessage(content)
	sms.SMSToolsFile = content

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
