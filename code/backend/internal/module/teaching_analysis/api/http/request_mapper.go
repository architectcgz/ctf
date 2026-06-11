package http

import teachinganalysisqueries "ctf-platform/internal/module/teaching_analysis/application/queries"

func toTeacherClassListInput(source TeacherClassQuery) *teachinganalysisqueries.TeacherClassListInput {
	return &teachinganalysisqueries.TeacherClassListInput{
		Page: source.Page,
		Size: source.Size,
	}
}

func toTeacherStudentDirectoryInput(source TeacherStudentDirectoryQuery) *teachinganalysisqueries.TeacherStudentDirectoryInput {
	return &teachinganalysisqueries.TeacherStudentDirectoryInput{
		ClassName: source.ClassName,
		Keyword:   source.Keyword,
		StudentNo: source.StudentNo,
		Page:      source.Page,
		Size:      source.Size,
		SortKey:   source.SortKey,
		SortOrder: source.SortOrder,
	}
}

func toTeacherStudentListInput(source TeacherStudentQuery) *teachinganalysisqueries.TeacherStudentListInput {
	return &teachinganalysisqueries.TeacherStudentListInput{
		Keyword:   source.Keyword,
		StudentNo: source.StudentNo,
	}
}

func toTeacherClassInsightInput(source TeacherClassInsightQuery) *teachinganalysisqueries.TeacherClassInsightInput {
	return &teachinganalysisqueries.TeacherClassInsightInput{
		FromDate: source.FromDate,
		ToDate:   source.ToDate,
	}
}

func toTeacherEvidenceInput(source TeacherEvidenceQuery) *teachinganalysisqueries.TeacherEvidenceInput {
	return &teachinganalysisqueries.TeacherEvidenceInput{
		ChallengeID: source.ChallengeID,
		ContestID:   source.ContestID,
		RoundID:     source.RoundID,
		EventType:   source.EventType,
		From:        source.From,
		To:          source.To,
		Limit:       source.Limit,
		Offset:      source.Offset,
	}
}

func toTeacherAttackSessionInput(source TeacherAttackSessionQuery) *teachinganalysisqueries.TeacherAttackSessionInput {
	return &teachinganalysisqueries.TeacherAttackSessionInput{
		Mode:        source.Mode,
		ChallengeID: source.ChallengeID,
		ContestID:   source.ContestID,
		RoundID:     source.RoundID,
		Result:      source.Result,
		WithEvents:  source.WithEvents,
		Limit:       source.Limit,
		Offset:      source.Offset,
	}
}
