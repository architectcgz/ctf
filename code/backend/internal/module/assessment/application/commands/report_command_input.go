package commands

type CreatePersonalReportInput struct {
	Format string
}

type CreateClassReportInput struct {
	ClassName string
	Format    string
}

type CreateContestExportInput struct {
	Format string
}

type CreateStudentReviewArchiveInput struct {
	Format string
}

type CreateTeacherAWDReviewExportInput struct {
	RoundNumber *int
}
