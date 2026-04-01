package dto

type CreatePersonalReportReq struct {
	Format string `json:"format" binding:"omitempty,oneof=pdf excel"`
}

type CreateClassReportReq struct {
	ClassName string `json:"class_name"`
	Format    string `json:"format" binding:"omitempty,oneof=pdf excel"`
}

type CreateContestExportReq struct {
	Format string `json:"format" binding:"omitempty,oneof=json"`
}

type CreateStudentReviewArchiveReq struct {
	Format string `json:"format" binding:"omitempty,oneof=json"`
}

type ReportExportData struct {
	ReportID     int64   `json:"report_id"`
	Status       string  `json:"status"`
	DownloadURL  *string `json:"download_url,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
}
