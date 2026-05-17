package http

import practiceports "ctf-platform/internal/module/practice/ports"

type practiceResponseMapperContract interface {
	ToProgressResp(source practiceports.UserProgressSnapshot) ProgressResp
	ToProgressRespPtr(source *practiceports.UserProgressSnapshot) *ProgressResp
	ToCategoryStat(source practiceports.UserProgressCategorySnapshot) CategoryStat
	ToDifficultyStat(source practiceports.UserProgressDifficultySnapshot) DifficultyStat
	ToTimelineResp(source practiceports.TimelineSnapshot) TimelineResp
	ToTimelineRespPtr(source *practiceports.TimelineSnapshot) *TimelineResp
	ToTimelineEvent(source practiceports.TimelineEventSnapshot) TimelineEvent
}

var practiceResponseMapper practiceResponseMapperContract

type practiceResponseMapperContractImpl struct{}

func (practiceResponseMapperContractImpl) ToProgressResp(source practiceports.UserProgressSnapshot) ProgressResp {
	resp := ProgressResp{
		TotalScore:      source.TotalScore,
		TotalSolved:     source.TotalSolved,
		Rank:            source.Rank,
		CategoryStats:   make([]CategoryStat, len(source.CategoryStats)),
		DifficultyStats: make([]DifficultyStat, len(source.DifficultyStats)),
	}
	for i, stat := range source.CategoryStats {
		resp.CategoryStats[i] = practiceResponseMapper.ToCategoryStat(stat)
	}
	for i, stat := range source.DifficultyStats {
		resp.DifficultyStats[i] = practiceResponseMapper.ToDifficultyStat(stat)
	}
	return resp
}

func (practiceResponseMapperContractImpl) ToProgressRespPtr(source *practiceports.UserProgressSnapshot) *ProgressResp {
	if source == nil {
		return nil
	}
	resp := practiceResponseMapper.ToProgressResp(*source)
	return &resp
}

func (practiceResponseMapperContractImpl) ToCategoryStat(source practiceports.UserProgressCategorySnapshot) CategoryStat {
	return CategoryStat{
		Category: source.Category,
		Solved:   source.Solved,
		Total:    source.Total,
	}
}

func (practiceResponseMapperContractImpl) ToDifficultyStat(source practiceports.UserProgressDifficultySnapshot) DifficultyStat {
	return DifficultyStat{
		Difficulty: source.Difficulty,
		Solved:     source.Solved,
		Total:      source.Total,
	}
}

func (practiceResponseMapperContractImpl) ToTimelineResp(source practiceports.TimelineSnapshot) TimelineResp {
	resp := TimelineResp{
		Events: make([]TimelineEvent, len(source.Events)),
	}
	for i, event := range source.Events {
		resp.Events[i] = practiceResponseMapper.ToTimelineEvent(event)
	}
	return resp
}

func (practiceResponseMapperContractImpl) ToTimelineRespPtr(source *practiceports.TimelineSnapshot) *TimelineResp {
	if source == nil {
		return nil
	}
	resp := practiceResponseMapper.ToTimelineResp(*source)
	return &resp
}

func (practiceResponseMapperContractImpl) ToTimelineEvent(source practiceports.TimelineEventSnapshot) TimelineEvent {
	return TimelineEvent{
		Type:        source.Type,
		ChallengeID: source.ChallengeID,
		Title:       source.Title,
		Timestamp:   source.Timestamp,
		IsCorrect:   source.IsCorrect,
		Points:      source.Points,
		Detail:      source.Detail,
	}
}
