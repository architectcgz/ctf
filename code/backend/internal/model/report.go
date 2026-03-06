package model

import "time"

const (
	ReportTypePersonal = "personal"
	ReportTypeClass    = "class"
)

const (
	ReportFormatPDF   = "pdf"
	ReportFormatExcel = "excel"
)

const (
	ReportStatusProcessing = "processing"
	ReportStatusReady      = "ready"
	ReportStatusFailed     = "failed"
)

type Report struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	Type        string     `gorm:"column:type"`
	Format      string     `gorm:"column:format"`
	UserID      *int64     `gorm:"column:user_id"`
	ClassName   *string    `gorm:"column:class_name"`
	Status      string     `gorm:"column:status"`
	FilePath    string     `gorm:"column:file_path"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	ErrorMsg    *string    `gorm:"column:error_msg"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
}

func (Report) TableName() string {
	return "reports"
}
