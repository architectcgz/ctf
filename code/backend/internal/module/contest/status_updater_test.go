package contest

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
)

type statusUpdaterRepoStub struct {
	contests       []*model.Contest
	updatedStatus  map[int64]string
	receivedStatus []string
}

func (s *statusUpdaterRepoStub) Create(context.Context, *model.Contest) error {
	panic("unexpected call")
}

func (s *statusUpdaterRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	panic("unexpected call")
}

func (s *statusUpdaterRepoStub) Update(context.Context, *model.Contest) error {
	panic("unexpected call")
}

func (s *statusUpdaterRepoStub) FindTeamsByIDs(context.Context, []int64) ([]*model.Team, error) {
	panic("unexpected call")
}

func (s *statusUpdaterRepoStub) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	panic("unexpected call")
}

func (s *statusUpdaterRepoStub) ListByStatusesAndTimeRange(_ context.Context, statuses []string, _ time.Time, _, _ int) ([]*model.Contest, int64, error) {
	s.receivedStatus = append([]string(nil), statuses...)
	return s.contests, int64(len(s.contests)), nil
}

func (s *statusUpdaterRepoStub) UpdateStatus(_ context.Context, id int64, status string) error {
	if s.updatedStatus == nil {
		s.updatedStatus = make(map[int64]string)
	}
	s.updatedStatus[id] = status
	return nil
}

func TestStatusUpdaterUpdateStatuses_EndsFrozenContest(t *testing.T) {
	now := time.Now()
	repo := &statusUpdaterRepoStub{
		contests: []*model.Contest{
			{
				ID:        7,
				Status:    model.ContestStatusFrozen,
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(-time.Minute),
			},
		},
	}
	updater := NewStatusUpdater(repo, nil, time.Minute, 100, nil)

	updater.updateStatuses(context.Background())

	if got := repo.updatedStatus[7]; got != model.ContestStatusEnded {
		t.Fatalf("expected frozen contest to end, got %q", got)
	}
}

func TestStatusUpdaterUpdateStatuses_RequestsFrozenStatus(t *testing.T) {
	repo := &statusUpdaterRepoStub{}
	updater := NewStatusUpdater(repo, nil, time.Minute, 100, nil)

	updater.updateStatuses(context.Background())

	expected := []string{
		model.ContestStatusRegistration,
		model.ContestStatusRunning,
		model.ContestStatusFrozen,
	}
	if len(repo.receivedStatus) != len(expected) {
		t.Fatalf("expected %d statuses, got %d", len(expected), len(repo.receivedStatus))
	}
	for i, status := range expected {
		if repo.receivedStatus[i] != status {
			t.Fatalf("expected status %q at index %d, got %q", status, i, repo.receivedStatus[i])
		}
	}
}
