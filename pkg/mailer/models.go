package mailer

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
	SMSToolsFile string `json:"smstools_file"`
}

type SMSEnqueueRequest struct {
	SMSMessage SMSMessage `json:"sms_message"`
	FolderName string     `json:"folder_name"`
}
