package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ctf-platform/internal/auditlog"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	fullrouteradmin "ctf-platform/tests/system/http/fullrouteradmin"
)

func TestFullRouter_AdminChallengeManagementStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	fullrouteradmin.VerifyAdminChallengeManagementStateMatrix(
		t,
		fullrouteradmin.ChallengeManagementDriver{
			Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			},
			AdminHeaders:            bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
			TeacherHeaders:          bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
			OtherTeacherHeaders:     bearerHeaders(loginForToken(t, env.router, env.otherTeacher.Username, "Password123")),
			StudentHeaders:          bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123")),
			ImageID:                 env.image.ID,
			PracticeStudentID:       env.peerStudent.ID,
			PracticeStudentUsername: env.peerStudent.Username,
			PublishChallenge: func(t *testing.T, challengeID int64) {
				if err := env.db.Model(&appChallengeRow{}).
					Where("id = ?", challengeID).
					Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
					t.Fatalf("set challenge %d published: %v", challengeID, err)
				}
			},
			CreatePracticeSubmission: func(t *testing.T, challengeID int64) {
				createPracticeSubmission(t, env, env.peerStudent.ID, challengeID, 150)
			},
			SetPracticeStudentNo: func(t *testing.T, studentNo string) {
				if err := env.db.Model(&identitycontracts.User{}).Where("id = ?", env.peerStudent.ID).Update("student_no", studentNo).Error; err != nil {
					t.Fatalf("set peer student number: %v", err)
				}
				env.peerStudent.StudentNo = studentNo
			},
			CreateDeleteBlockedChallenge: func(t *testing.T, title string) int64 {
				return createDraftChallengeRecord(t, env, title).ID
			},
			CreateRunningInstanceForDeleteBlock: func(t *testing.T, challengeID int64) {
				createRunningInstanceForChallenge(t, env, challengeID, env.student.ID)
			},
			StopInstancesForChallenge: func(t *testing.T, challengeID int64) {
				stopInstancesForChallenge(t, env, challengeID)
			},
		},
	)
}

func TestFullRouter_AdminOpsAndNotificationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	fullrouteradmin.VerifyAdminOpsAndNotificationStateMatrix(
		t,
		fullrouteradmin.AdminOpsAndNotificationDriver{
			Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			},
			MultipartRequest: func(method, target, fieldName, fileName, content string, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterMultipartRequest(t, env.router, method, target, fieldName, fileName, content, headers)
			},
			Router:         env.router,
			AdminHeaders:   bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
			TeacherHeaders: bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
			StudentHeaders: bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
			PeerHeaders:    bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123")),
			SeededImageID:  env.image.ID,
			NotificationID: env.notification.ID,
			SeedOnlineSession: func(t *testing.T) {
				if err := env.cache.Set(context.Background(), "ctf:auth:session:manual-online", "online", time.Hour).Err(); err != nil {
					t.Fatalf("seed session key: %v", err)
				}
			},
			SeedAuditLogs: func(t *testing.T) {
				submitDetail, _ := json.Marshal(map[string]any{"username": env.student.Username, "source": "matrix"})
				for i := 0; i < 5; i++ {
					if err := env.db.Create(&opsentity.AuditLog{
						UserID:       &env.student.ID,
						Action:       auditlog.ActionSubmit,
						ResourceType: "challenge_submission",
						Detail:       string(submitDetail),
						IPAddress:    "10.0.0.1",
						CreatedAt:    time.Now().Add(-time.Duration(i) * time.Minute),
					}).Error; err != nil {
						t.Fatalf("seed submit audit log: %v", err)
					}
				}
				for _, user := range []*identitycontracts.User{env.student, env.peerStudent} {
					if err := env.db.Create(&opsentity.AuditLog{
						UserID:       &user.ID,
						Action:       auditlog.ActionLogin,
						ResourceType: "auth_login",
						Detail:       `{"username":"` + user.Username + `"}`,
						IPAddress:    "10.0.0.99",
						CreatedAt:    time.Now().Add(-10 * time.Minute),
					}).Error; err != nil {
						t.Fatalf("seed login audit log: %v", err)
					}
				}
			},
		},
	)
}

func TestFullRouter_AdminImagesCapsOversizedPageSize(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/images?page=1&page_size=200", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var payload struct {
		List []challengehttp.ImageResp `json:"list"`
		Page int                       `json:"page"`
		Size int                       `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &payload)

	if payload.Page != 1 {
		t.Fatalf("expected page=1, got %d", payload.Page)
	}
	if payload.Size != 100 {
		t.Fatalf("expected capped page_size=100, got %d", payload.Size)
	}
	if len(payload.List) == 0 {
		t.Fatal("expected image list to contain seeded records")
	}
}
