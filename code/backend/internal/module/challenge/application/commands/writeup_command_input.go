package commands

type UpsertOfficialWriteupInput struct {
	Title      string
	Content    string
	Visibility string
}

type UpsertSubmissionWriteupInput struct {
	Title            string
	Content          string
	SubmissionStatus string
}
