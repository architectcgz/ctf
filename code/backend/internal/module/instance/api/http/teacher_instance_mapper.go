package http

import instancecontracts "ctf-platform/internal/module/instance/contracts"

func mapTeacherInstanceListQuery(source TeacherInstanceQuery) instancecontracts.TeacherInstanceListQuery {
	return instancecontracts.TeacherInstanceListQuery{
		ClassName: source.ClassName,
		Keyword:   source.Keyword,
		StudentNo: source.StudentNo,
		Status:    source.Status,
		Page:      source.Page,
		PageSize:  source.PageSize,
	}
}

func mapTeacherInstanceItems(source []instancecontracts.TeacherInstanceItem) []TeacherInstanceItem {
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

func mapTeacherInstancePageResp(source *instancecontracts.TeacherInstancePageResult) *TeacherInstancePageResp {
	if source == nil {
		return &TeacherInstancePageResp{
			List:     []TeacherInstanceItem{},
			Page:     1,
			PageSize: 20,
		}
	}

	return &TeacherInstancePageResp{
		List:     mapTeacherInstanceItems(source.List),
		Total:    source.Total,
		Page:     source.Page,
		PageSize: source.PageSize,
		Summary: TeacherInstanceListSummaryResp{
			TotalCount:        source.Summary.TotalCount,
			RunningCount:      source.Summary.RunningCount,
			ExpiringSoonCount: source.Summary.ExpiringSoonCount,
			WarningCount:      source.Summary.WarningCount,
		},
	}
}
