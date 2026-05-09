package queries

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type contestListRepoStub struct{}

func (s *contestListRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

func (s *contestListRepoStub) List(context.Context, contestports.ContestListFilter, int, int) ([]*model.Contest, int64, error) {
	return []*model.Contest{}, 0, nil
}

func (s *contestListRepoStub) Summarize(context.Context, contestports.ContestListFilter) (contestports.ContestListSummary, error) {
	return contestports.ContestListSummary{}, nil
}

type contestListRepoSpy struct {
	listFilter    contestports.ContestListFilter
	listOffset    int
	listLimit     int
	summaryFilter contestports.ContestListFilter
}

func (s *contestListRepoSpy) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

func (s *contestListRepoSpy) List(_ context.Context, filter contestports.ContestListFilter, offset, limit int) ([]*model.Contest, int64, error) {
	s.listFilter = filter
	s.listOffset = offset
	s.listLimit = limit
	return []*model.Contest{}, 0, nil
}

func (s *contestListRepoSpy) Summarize(_ context.Context, filter contestports.ContestListFilter) (contestports.ContestListSummary, error) {
	s.summaryFilter = filter
	return contestports.ContestListSummary{
		RegistrationCount: 2,
		RunningCount:      3,
		FrozenCount:       1,
		EndedCount:        4,
	}, nil
}

func TestContestServiceGetContestReturnsContestNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestService(&contestListRepoStub{}, zap.NewNop())

	_, err := service.GetContest(context.Background(), 42)
	if err != errcode.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}

func TestContestServiceListContestsAcceptsNarrowRepository(t *testing.T) {
	t.Parallel()

	service := NewContestService(&contestListRepoStub{}, zap.NewNop())

	items, total, err := service.ListContests(context.Background(), ListContestsInput{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListContests() error = %v", err)
	}
	if len(items) != 0 || total != 0 {
		t.Fatalf("unexpected list result: items=%d total=%d", len(items), total)
	}
}

func TestContestServiceListContestsNormalizesSortAndKeepsFilterContract(t *testing.T) {
	t.Parallel()

	repo := &contestListRepoSpy{}
	service := NewContestService(repo, zap.NewNop())

	_, _, err := service.ListContests(context.Background(), ListContestsInput{
		Statuses:  []string{"running", "ended"},
		SortKey:   "start_time",
		SortOrder: "asc",
		Page:      2,
		Size:      20,
	})
	if err != nil {
		t.Fatalf("ListContests() error = %v", err)
	}

	if repo.listOffset != 20 || repo.listLimit != 20 {
		t.Fatalf("unexpected paging: offset=%d limit=%d", repo.listOffset, repo.listLimit)
	}
	if got, want := contestports.ContestListFilterStatuses(repo.listFilter), []string{"running", "ended"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("unexpected statuses: got=%v want=%v", got, want)
	}
	if sort := contestports.ContestListFilterSort(repo.listFilter); !contestports.ContestListSortIsStartTime(sort) || !contestports.ContestListSortIsAsc(sort) {
		t.Fatalf("unexpected sort filter: %+v", repo.listFilter)
	}
}

func TestContestServiceGetContestListSummaryUsesNormalizedFilter(t *testing.T) {
	t.Parallel()

	repo := &contestListRepoSpy{}
	service := NewContestService(repo, zap.NewNop())

	summary, err := service.GetContestListSummary(context.Background(), ListContestsInput{
		Statuses:  []string{"registration", "running"},
		SortKey:   "invalid",
		SortOrder: "invalid",
	})
	if err != nil {
		t.Fatalf("GetContestListSummary() error = %v", err)
	}

	if got, want := contestports.ContestListFilterStatuses(repo.summaryFilter), []string{"registration", "running"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("unexpected summary statuses: got=%v want=%v", got, want)
	}
	if sort := contestports.ContestListFilterSort(repo.summaryFilter); contestports.ContestListSortIsStartTime(sort) || contestports.ContestListSortIsAsc(sort) {
		t.Fatalf("unexpected summary sort filter: %+v", repo.summaryFilter)
	}
	if summary.RegistrationCount != 2 || summary.RunningCount != 3 || summary.FrozenCount != 1 || summary.EndedCount != 4 {
		t.Fatalf("unexpected summary payload: %+v", summary)
	}
}
