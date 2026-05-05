package commands

import (
	"testing"
	"time"

	"ctf-platform/internal/model"
)

func TestContestRespFromModelReturnsTimesInUTC(t *testing.T) {
	shanghai := time.FixedZone("Asia/Shanghai", 8*60*60)
	start := time.Date(2026, 4, 28, 12, 36, 54, 0, shanghai)
	end := start.Add(2 * time.Hour)
	freeze := end.Add(-30 * time.Minute)
	created := start.Add(-time.Hour)
	updated := created.Add(time.Minute)

	resp := contestRespFromModel(&model.Contest{
		ID:         6,
		Title:      "time contract",
		Mode:       model.ContestModeAWD,
		StartTime:  start,
		EndTime:    end,
		FreezeTime: &freeze,
		Status:     model.ContestStatusRunning,
		CreatedAt:  created,
		UpdatedAt:  updated,
	})

	if resp.StartTime.Location() != time.UTC || resp.EndTime.Location() != time.UTC {
		t.Fatalf("expected UTC contest window, got start=%v end=%v", resp.StartTime.Location(), resp.EndTime.Location())
	}
	if resp.FreezeTime == nil || resp.FreezeTime.Location() != time.UTC {
		t.Fatalf("expected UTC freeze time, got %+v", resp.FreezeTime)
	}
	if resp.CreatedAt.Location() != time.UTC || resp.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UTC metadata times, got created=%v updated=%v", resp.CreatedAt.Location(), resp.UpdatedAt.Location())
	}
	if !resp.StartTime.Equal(start) || !resp.EndTime.Equal(end) || !resp.FreezeTime.Equal(freeze) {
		t.Fatalf("response changed instant: start=%s end=%s freeze=%s", resp.StartTime, resp.EndTime, *resp.FreezeTime)
	}
}
