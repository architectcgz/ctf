package app

import (
	"net/http/httptest"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"ctf-platform/internal/shared/taxonomy"
	fullrouteradmin "ctf-platform/tests/system/http/fullrouteradmin"
)

func TestFullRouter_AdminCanToggleAWDControlsAndSeeOrchestrationState(t *testing.T) {
	env := newFullRouterTestEnv(t)
	now := time.Now().UTC()

	awdTeam := &contestcontracts.Team{
		ContestID:  env.awdContest.ID,
		Name:       "AWD Control Team",
		CaptainID:  env.student.ID,
		InviteCode: "AWDCTRL1",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.db.Create(awdTeam).Error; err != nil {
		t.Fatalf("create awd team: %v", err)
	}
	if err := env.db.Create(&contestcontracts.TeamMember{
		ContestID: env.awdContest.ID,
		TeamID:    awdTeam.ID,
		UserID:    env.student.ID,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create awd team member: %v", err)
	}

	serviceSnapshot, err := contestcontracts.EncodeContestAWDServiceSnapshot(contestcontracts.ContestAWDServiceSnapshot{
		Name:       "AWD Web",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		RuntimeConfig: map[string]any{
			"image_id":         env.image.ID,
			"instance_sharing": challengecontracts.InstanceSharingPerTeam,
		},
		FlagConfig: map[string]any{
			"flag_type":   challengecontracts.FlagTypeStatic,
			"flag_prefix": "flag",
		},
	})
	if err != nil {
		t.Fatalf("encode awd service snapshot: %v", err)
	}
	awdService := &contestcontracts.ContestAWDService{
		ContestID:       env.awdContest.ID,
		AWDChallengeID:  env.challenge.ID,
		DisplayName:     "AWD Web",
		IsVisible:       true,
		ServiceSnapshot: serviceSnapshot,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := env.db.Create(awdService).Error; err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	adminHeaders := sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))
	fullrouteradmin.VerifyAdminCanToggleAWDControlsAndSeeOrchestrationState(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		adminHeaders,
		fullrouteradmin.AWDControlTarget{
			ContestID: env.awdContest.ID,
			TeamID:    awdTeam.ID,
			ServiceID: awdService.ID,
		},
	)
}

func TestFullRouter_AdminChallengePublishRequestLifecycle(t *testing.T) {
	env := newFullRouterTestEnv(t)
	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", env.challenge.ID).
		Update("status", challengecontracts.ChallengeStatusDraft).Error; err != nil {
		t.Fatalf("set challenge draft: %v", err)
	}
	env.challenge.Status = challengecontracts.ChallengeStatusDraft

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	fullrouteradmin.VerifyAdminChallengePublishRequestLifecycle(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		teacherHeaders,
		fullrouteradmin.PublishRequestLifecycleTarget{
			ChallengeID: env.challenge.ID,
		},
	)
}
