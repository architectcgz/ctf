package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

type ContestWriteRepository interface {
	Create(ctx context.Context, contest *model.Contest) error
	Update(ctx context.Context, contest *model.Contest) error
}

type ContestLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Contest, error)
}

type contestListSortKey uint8

const (
	contestListSortKeyCreatedAt contestListSortKey = iota
	contestListSortKeyStartTime
)

type contestListSortOrder uint8

const (
	contestListSortOrderDesc contestListSortOrder = iota
	contestListSortOrderAsc
)

type ContestListSort interface {
	isContestListSort()
}

type contestListSort struct {
	key   contestListSortKey
	order contestListSortOrder
}

func NewContestListSortByCreatedAtDesc() ContestListSort {
	return contestListSort{
		key:   contestListSortKeyCreatedAt,
		order: contestListSortOrderDesc,
	}
}

func NewContestListSortByCreatedAtAsc() ContestListSort {
	return contestListSort{
		key:   contestListSortKeyCreatedAt,
		order: contestListSortOrderAsc,
	}
}

func NewContestListSortByStartTimeDesc() ContestListSort {
	return contestListSort{
		key:   contestListSortKeyStartTime,
		order: contestListSortOrderDesc,
	}
}

func NewContestListSortByStartTimeAsc() ContestListSort {
	return contestListSort{
		key:   contestListSortKeyStartTime,
		order: contestListSortOrderAsc,
	}
}

func (contestListSort) isContestListSort() {}

func ContestListSortIsStartTime(sort ContestListSort) bool {
	return mustContestListSort(sort).key == contestListSortKeyStartTime
}

func ContestListSortIsAsc(sort ContestListSort) bool {
	return mustContestListSort(sort).order == contestListSortOrderAsc
}

type ContestListFilter interface {
	isContestListFilter()
}

type contestListFilter struct {
	statuses []string
	mode     *string
	sort     ContestListSort
}

func NewContestListFilter(statuses []string, mode *string, sort ContestListSort) ContestListFilter {
	if sort == nil {
		panic("contest list sort is required")
	}
	return contestListFilter{
		statuses: statuses,
		mode:     mode,
		sort:     sort,
	}
}

func (contestListFilter) isContestListFilter() {}

func ContestListFilterStatuses(filter ContestListFilter) []string {
	return mustContestListFilter(filter).statuses
}

func ContestListFilterMode(filter ContestListFilter) *string {
	return mustContestListFilter(filter).mode
}

func ContestListFilterSort(filter ContestListFilter) ContestListSort {
	return mustContestListFilter(filter).sort
}

func mustContestListFilter(filter ContestListFilter) contestListFilter {
	concrete, ok := filter.(contestListFilter)
	if !ok {
		panic("invalid contest list filter")
	}
	return concrete
}

func mustContestListSort(sort ContestListSort) contestListSort {
	concrete, ok := sort.(contestListSort)
	if !ok {
		panic("invalid contest list sort")
	}
	return concrete
}

type ContestListSummary struct {
	DraftCount        int64
	RegistrationCount int64
	RunningCount      int64
	FrozenCount       int64
	EndedCount        int64
}

type ContestListRepository interface {
	ContestLookupRepository
	List(ctx context.Context, filter ContestListFilter, offset, limit int) ([]*model.Contest, int64, error)
	Summarize(ctx context.Context, filter ContestListFilter) (ContestListSummary, error)
}

type ContestScoreboardRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Contest, error)
	FindTeamsByIDs(ctx context.Context, ids []int64) ([]*model.Team, error)
	FindScoreboardTeamStats(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) (map[int64]ScoreboardTeamStats, error)
}

type ContestScoreboardAdminRepository interface {
	ContestLookupRepository
	Update(ctx context.Context, contest *model.Contest) error
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
}

type ScoreboardMemberScore struct {
	Member string
	Score  float64
}

type ScoreboardTeamScoreEntry struct {
	TeamID int64
	Score  float64
}

type ScoreboardTeamRank struct {
	Rank  int
	Score float64
}

type ContestScoreboardStateStore interface {
	HasFrozenScoreboardSnapshot(ctx context.Context, contestID int64) (bool, error)
	CreateFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error
	ClearFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error
	ListLiveScoreboard(ctx context.Context, contestID int64) ([]ScoreboardMemberScore, error)
	ListFrozenScoreboard(ctx context.Context, contestID int64) ([]ScoreboardMemberScore, error)
	GetLiveTeamRank(ctx context.Context, contestID, teamID int64) (ScoreboardTeamRank, bool, error)
	IncrementLiveTeamScore(ctx context.Context, contestID, teamID int64, points float64) error
	ReplaceLiveScoreboard(ctx context.Context, contestID int64, entries []ScoreboardTeamScoreEntry) error
}

type ContestStatusRepository interface {
	ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error)
	ApplyStatusTransition(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error)
}

type ContestEndedRuntimeCleaner interface {
	CleanupEndedContestAWDInstances(ctx context.Context, contestID int64) error
}

type ContestStatusSideEffectStore interface {
	CreateFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error
	ClearFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error
	ClearEndedContestRuntimeState(ctx context.Context, contestID int64) error
}

type ContestSchedulerLockLease interface {
	Key(ctx context.Context) string
	Refresh(ctx context.Context, ttl time.Duration) (bool, error)
	Release(ctx context.Context) (bool, error)
}

type ContestStatusUpdateLockStore interface {
	AcquireStatusUpdateLock(ctx context.Context, ttl time.Duration) (ContestSchedulerLockLease, bool, error)
}

type AWDServiceStatusEntry struct {
	TeamID    int64
	ServiceID int64
	Status    string
}

type AWDRoundStateStore interface {
	AcquireAWDSchedulerLock(ctx context.Context, ttl time.Duration) (ContestSchedulerLockLease, bool, error)
	TryAcquireAWDRoundLock(ctx context.Context, contestID int64, roundNumber int, ttl time.Duration) (bool, error)
	IsAWDCurrentRound(ctx context.Context, contestID int64, roundNumber int) (bool, error)
	LoadAWDCurrentRoundNumber(ctx context.Context, contestID int64) (int, bool, error)
	LoadAWDRoundFlag(ctx context.Context, contestID, roundID, teamID, serviceID int64) (string, bool, error)
	SyncAWDCurrentRoundState(ctx context.Context, contestID int64, round *model.AWDRound, assignments []AWDFlagAssignment, ttl time.Duration) error
	ClearAWDCurrentRoundState(ctx context.Context, contestID int64) error
	SetAWDServiceStatus(ctx context.Context, contestID, teamID, serviceID int64, status string) error
	ReplaceAWDServiceStatus(ctx context.Context, contestID int64, entries []AWDServiceStatusEntry) error
	ClearAWDServiceStatus(ctx context.Context, contestID int64) error
}
