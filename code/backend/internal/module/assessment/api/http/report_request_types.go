package http

type CreatePersonalReportReq struct {
	Format string `json:"format" binding:"omitempty,oneof=pdf excel"`
}

type CreateClassReportReq struct {
	ClassName string `json:"class_name"`
	Format    string `json:"format" binding:"omitempty,oneof=pdf excel"`
	FromDate  string `json:"from_date,omitempty"`
	ToDate    string `json:"to_date,omitempty"`
}

type CreateContestExportReq struct {
	Format string `json:"format" binding:"omitempty,oneof=json"`
}

type CreateStudentReviewArchiveReq struct {
	Format string `json:"format" binding:"omitempty,oneof=json"`
}

type CreateTeacherAWDReviewExportReq struct {
	RoundNumber *int `json:"round_number,omitempty"`
}
