package app

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	"ctf-platform/internal/shared/taxonomy"
	fullrouterteacherauthoring "ctf-platform/tests/system/http/fullrouterteacherauthoring"
)

func TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	fullrouterteacherauthoring.VerifyTeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges(
		t,
		fullrouterteacherauthoring.ChallengeOwnershipDriver{
			Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			},
			AdminHeaders:   adminHeaders,
			TeacherHeaders: teacherHeaders,
			CreatePayload: func(title string) map[string]any {
				return map[string]any{
					"title":       title,
					"description": "ownership test challenge",
					"category":    taxonomy.DimensionWeb,
					"difficulty":  taxonomy.DifficultyEasy,
					"points":      100,
					"image_id":    env.image.ID,
				}
			},
			TemplateID: env.template.ID,
			ImageID:    env.image.ID,
			PrepareArchivedAdminChallenge: func(t *testing.T, adminChallengeID int64) int64 {
				return prepareArchivedAdminChallengePackageExport(t, env, adminChallengeID)
			},
		},
	)
}

func TestFullRouter_CreateChallengeStoresCreator(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	challengeID := fullrouterteacherauthoring.VerifyCreateChallengeStoresCreatorResponse(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		teacherHeaders,
		map[string]any{
			"title":       "creator-marker",
			"description": "creator marker challenge",
			"category":    taxonomy.DimensionWeb,
			"difficulty":  taxonomy.DifficultyEasy,
			"points":      100,
			"image_id":    env.image.ID,
		},
	)

	var createdBy sql.NullInt64
	if err := env.db.Raw("SELECT created_by FROM challenges WHERE id = ?", challengeID).Scan(&createdBy).Error; err != nil {
		t.Fatalf("query challenge created_by: %v", err)
	}
	if !createdBy.Valid || createdBy.Int64 != env.teacher.ID {
		t.Fatalf("unexpected created_by=%+v, want %d", createdBy, env.teacher.ID)
	}
}

func TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	fullrouterteacherauthoring.VerifyChallengeSelfCheckRunsPrecheckAndRuntime(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		teacherHeaders,
		env.challenge.ID,
	)
}

func prepareArchivedAdminChallengePackageExport(t *testing.T, env *fullRouterTestEnv, adminChallengeID int64) int64 {
	t.Helper()

	packageArchiveDir := filepath.Join(t.TempDir(), "package-export")
	if err := os.MkdirAll(packageArchiveDir, 0o755); err != nil {
		t.Fatalf("mkdir package export dir: %v", err)
	}
	packageArchivePath := filepath.Join(packageArchiveDir, "challenge-package.zip")
	if err := os.WriteFile(packageArchivePath, []byte("package export"), 0o644); err != nil {
		t.Fatalf("write package export: %v", err)
	}
	now := time.Now().UTC()
	packageRevision := &challengeentity.ChallengePackageRevision{
		ChallengeID:      adminChallengeID,
		RevisionNo:       1,
		SourceType:       challengeentity.ChallengePackageRevisionSourceExported,
		ArchivePath:      packageArchivePath,
		SourceDir:        filepath.Join(packageArchiveDir, "source"),
		ManifestSnapshot: "{}",
		TopologySnapshot: "{}",
		CreatedBy:        &env.admin.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := os.MkdirAll(packageRevision.SourceDir, 0o755); err != nil {
		t.Fatalf("mkdir package source dir: %v", err)
	}
	if err := env.db.Create(packageRevision).Error; err != nil {
		t.Fatalf("create package revision: %v", err)
	}

	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", adminChallengeID).
		Update("status", challengecontracts.ChallengeStatusArchived).Error; err != nil {
		t.Fatalf("archive admin challenge: %v", err)
	}

	return packageRevision.ID
}
