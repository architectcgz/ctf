package commands

type ReportExportData struct {
	ReportID     int64   `json:"report_id"`
	Status       string  `json:"status"`
	DownloadURL  *string `json:"download_url,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
}
