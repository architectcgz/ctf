package domain

import (
	"testing"

	"ctf-platform/internal/model"
)

func TestShouldGateAWDContestStart(t *testing.T) {
	running := model.ContestStatusRunning

	tests := []struct {
		name          string
		mode          string
		currentStatus string
		targetStatus  *string
		want          bool
	}{
		{
			name:          "gate awd start transition",
			mode:          model.ContestModeAWD,
			currentStatus: model.ContestStatusRegistration,
			targetStatus:  &running,
			want:          true,
		},
		{
			name:          "skip non awd contest",
			mode:          model.ContestModeJeopardy,
			currentStatus: model.ContestStatusRegistration,
			targetStatus:  &running,
			want:          false,
		},
		{
			name:          "skip when already running",
			mode:          model.ContestModeAWD,
			currentStatus: model.ContestStatusRunning,
			targetStatus:  &running,
			want:          false,
		},
		{
			name:          "skip other target status",
			mode:          model.ContestModeAWD,
			currentStatus: model.ContestStatusRegistration,
			targetStatus:  strPtr(model.ContestStatusFrozen),
			want:          false,
		},
		{
			name:          "skip nil target status",
			mode:          model.ContestModeAWD,
			currentStatus: model.ContestStatusRegistration,
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
