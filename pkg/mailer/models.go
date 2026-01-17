package mailer

type SMSMessage struct {
	From      string `json:"from"`
	Sent      string `json:"sent"`
	Received  string `json:"received"`
	Message   string `json:"message"`
	IMEI      string `json:"imei"`
	IMSI      string `json:"imsi"`
	Length    int    `json:"length"`
	Alphabet  string `json:"alphabet"`

	// GSM modem specific fields
	ModemDevice string `json:"modem_device"` // e.g. ttyUSB0
	Index       int    `json:"index"`        // Index for deletion
	Modem       string `json:"modem,omitempty"` // Set to ModemDevice for backward compatibility
}

type SMSEnqueueRequest struct {
	SMSMessage SMSMessage `json:"sms_message"`
	FolderName string     `json:"folder_name"`
}
