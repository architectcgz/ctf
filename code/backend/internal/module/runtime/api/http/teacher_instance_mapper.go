package http

import instancecontracts "ctf-platform/internal/module/instance/contracts"

func toTeacherInstanceListQuery(source TeacherInstanceQuery) instancecontracts.TeacherInstanceListQuery {
	return instancecontracts.TeacherInstanceListQuery{
		ClassName: source.ClassName,
		Keyword:   source.Keyword,
		StudentNo: source.StudentNo,
	}
}

func toTeacherInstanceItems(source []instancecontracts.TeacherInstanceItem) []TeacherInstanceItem {
	items := make([]TeacherInstanceItem, len(source))
	for idx, item := range source {
		items[idx] = TeacherInstanceItem{
			ID:              item.ID,
			StudentID:       item.StudentID,
			StudentName:     item.StudentName,
			StudentUsername: item.StudentUsername,
			StudentNo:       item.StudentNo,
			ClassName:       item.ClassName,
			ChallengeID:     item.ChallengeID,
			ChallengeTitle:  item.ChallengeTitle,
			Status:          item.Status,
			AccessURL:       item.AccessURL,
			Access:          item.Access,
			ExpiresAt:       item.ExpiresAt,
			RemainingTime:   item.RemainingTime,
			ExtendCount:     item.ExtendCount,
			MaxExtends:      item.MaxExtends,
			CreatedAt:       item.CreatedAt,
		}
	}
	return items
}
