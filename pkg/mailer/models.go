package mailer

type SMSMessage struct {
	// Fields maintained from SMSTools3 version
	From      string `json:"from"`
	Sent      string `json:"sent"`
	Received  string `json:"received"`
	Message   string `json:"message"`
	IMEI      string `json:"imei"`
	IMSI      string `json:"imsi"`
	Length    int    `json:"length"`
	Alphabet  string `json:"alphabet"`

	// New fields for AT modem
	ModemDevice string `json:"modem_device"` // e.g. ttyUSB0
	Index       int    `json:"index"`        // Index for deletion
	Modem       string `json:"modem,omitempty"` // Set to ModemDevice for backward compatibility

	// Fields removed (SMSTools3 specific) - kept for backward compatibility
	// but no longer populated
	FromTOA      string `json:"from_toa,omitempty"`
	FromSMSC     string `json:"from_smsc,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Report       string `json:"report,omitempty"`
	SMSToolsFile string `json:"smstools_file,omitempty"`
}

type SMSEnqueueRequest struct {
	SMSMessage SMSMessage `json:"sms_message"`
	FolderName string     `json:"folder_name"`
}
