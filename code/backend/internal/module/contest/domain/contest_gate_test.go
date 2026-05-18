package domain

import (
	"testing"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func TestShouldGateAWDContestStart(t *testing.T) {
	running := contestentity.ContestStatusRunning

	tests := []struct {
		name          string
		mode          string
		currentStatus string
		targetStatus  *string
		want          bool
	}{
		{
			name:          "gate awd start transition",
			mode:          contestentity.ContestModeAWD,
			currentStatus: contestentity.ContestStatusRegistration,
			targetStatus:  &running,
			want:          true,
		},
		{
			name:          "skip non awd contest",
			mode:          contestentity.ContestModeJeopardy,
			currentStatus: contestentity.ContestStatusRegistration,
			targetStatus:  &running,
			want:          false,
		},
		{
			name:          "skip when already running",
			mode:          contestentity.ContestModeAWD,
			currentStatus: contestentity.ContestStatusRunning,
			targetStatus:  &running,
			want:          false,
		},
		{
			name:          "skip other target status",
			mode:          contestentity.ContestModeAWD,
			currentStatus: contestentity.ContestStatusRegistration,
			targetStatus:  strPtr(contestentity.ContestStatusFrozen),
			want:          false,
		},
		{
			name:          "skip nil target status",
			mode:          contestentity.ContestModeAWD,
			currentStatus: contestentity.ContestStatusRegistration,
			targetStatus:  nil,
			want:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldGateAWDContestStart(tc.mode, tc.currentStatus, tc.targetStatus); got != tc.want {
				t.Fatalf("ShouldGateAWDContestStart() = %v, want %v", got, tc.want)
			}
		})
	}
}

func strPtr(value string) *string {
	return &value
}
