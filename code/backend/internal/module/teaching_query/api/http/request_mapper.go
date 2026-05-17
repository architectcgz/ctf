package http

import teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"

func toTeacherClassListInput(source TeacherClassQuery) *teachingqueryqueries.TeacherClassListInput {
	return &teachingqueryqueries.TeacherClassListInput{
		Page: source.Page,
		Size: source.Size,
	}
}

func toTeacherStudentDirectoryInput(source TeacherStudentDirectoryQuery) *teachingqueryqueries.TeacherStudentDirectoryInput {
	return &teachingqueryqueries.TeacherStudentDirectoryInput{
		ClassName: source.ClassName,
		Keyword:   source.Keyword,
		StudentNo: source.StudentNo,
		Page:      source.Page,
		Size:      source.Size,
		SortKey:   source.SortKey,
		SortOrder: source.SortOrder,
	}
}

func toTeacherStudentListInput(source TeacherStudentQuery) *teachingqueryqueries.TeacherStudentListInput {
	return &teachingqueryqueries.TeacherStudentListInput{
		Keyword:   source.Keyword,
		StudentNo: source.StudentNo,
	}
}

func toTeacherClassInsightInput(source TeacherClassInsightQuery) *teachingqueryqueries.TeacherClassInsightInput {
	return &teachingqueryqueries.TeacherClassInsightInput{
		FromDate: source.FromDate,
		ToDate:   source.ToDate,
	}
}

func toTeacherEvidenceInput(source TeacherEvidenceQuery) *teachingqueryqueries.TeacherEvidenceInput {
	return &teachingqueryqueries.TeacherEvidenceInput{
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

func toTeacherAttackSessionInput(source TeacherAttackSessionQuery) *teachingqueryqueries.TeacherAttackSessionInput {
	return &teachingqueryqueries.TeacherAttackSessionInput{
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
