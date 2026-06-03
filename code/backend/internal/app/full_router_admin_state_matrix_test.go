package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/auditlog"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsentity "ctf-platform/internal/module/ops/entity"
	fullrouteradmin "ctf-platform/tests/system/http/fullrouteradmin"
	xws "golang.org/x/net/websocket"
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

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "integration image",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var freeImage challengehttp.ImageResp
	decodeFullRouterData(t, resp, &freeImage)
	if freeImage.Name != "matrix/webapp" {
		t.Fatalf("unexpected created image: %+v", freeImage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "duplicate image",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/images?name=matrix/status=available", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), map[string]any{
		"description": "updated image",
		"status":      "failed",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var loadedImage challengehttp.ImageResp
	decodeFullRouterData(t, resp, &loadedImage)
	if loadedImage.Status != "failed" || loadedImage.Description != "updated image" {
		t.Fatalf("unexpected loaded image: %+v", loadedImage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/images/%d", env.image.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/users?role=student&class_name=ClassA", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var userPage map[string]any
	decodeFullRouterData(t, resp, &userPage)
	if int(userPage["total"].(float64)) < 2 {
		t.Fatalf("expected student page results, got %+v", userPage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username":   "admin_created_student",
		"name":       "Created Student",
		"password":   "Password123",
		"email":      "created_student@example.com",
		"student_no": "20260001",
		"class_name": "ClassA",
		"role":       identitycontracts.RoleStudent,
		"status":     identitycontracts.UserStatusActive,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdUserWrap map[string]json.RawMessage
	decodeFullRouterData(t, resp, &createdUserWrap)
	createdUser := decodeFullRouterJSON[identityhttp.AdminUserResp](t, createdUserWrap["user"])
	if createdUser.Username != "admin_created_student" {
		t.Fatalf("unexpected created user: %+v", createdUser)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "admin_created_student",
		"password": "Password123",
		"role":     identitycontracts.RoleStudent,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	updatedTeacherNo := "T-9001"
	updatedRole := identitycontracts.RoleTeacher
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), map[string]any{
		"role":       updatedRole,
		"teacher_no": updatedTeacherNo,
		"class_name": "ClassTeach",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var updatedUserWrap map[string]json.RawMessage
	decodeFullRouterData(t, resp, &updatedUserWrap)
	updatedUser := decodeFullRouterJSON[identityhttp.AdminUserResp](t, updatedUserWrap["user"])
	if updatedUser.TeacherNo == nil || *updatedUser.TeacherNo != updatedTeacherNo || updatedUser.StudentNo != nil {
		t.Fatalf("unexpected updated user: %+v", updatedUser)
	}

	csvContent := strings.Join([]string{
		"username,password,email,class_name,role,status,student_no,teacher_no,name",
		"import_new,Password123,import_new@example.com,ClassA,student,active,20260002,,Import New",
		"admin_created_student,,updated_import@example.com,ClassTeach,teacher,inactive,,T-9002,Imported Update",
		",Password123,bad@example.com,ClassA,student,active,20260003,,Bad Row",
	}, "\n")
	resp = performFullRouterMultipartRequest(t, env.router, http.MethodPost, "/api/v1/admin/users/import", "file", "users.csv", csvContent, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusCreated)

	var importResult identityhttp.ImportUsersResp
	decodeFullRouterData(t, resp, &importResult)
	if importResult.Created != 1 || importResult.Updated != 1 || importResult.Failed != 1 {
		t.Fatalf("unexpected import result: %+v", importResult)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	if err := env.cache.Set(context.Background(), "ctf:auth:session:manual-online", "online", time.Hour).Err(); err != nil {
		t.Fatalf("seed session key: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/dashboard", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var dashboard opshttp.DashboardStats
	decodeFullRouterData(t, resp, &dashboard)
	if dashboard.OnlineUsers < 1 || dashboard.ActiveContainers < 1 {
		t.Fatalf("unexpected dashboard stats: %+v", dashboard)
	}

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

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/audit-logs?action=submit&page=1&page_size=10", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var auditPage map[string]any
	decodeFullRouterData(t, resp, &auditPage)
	if int(auditPage["total"].(float64)) < 5 {
		t.Fatalf("unexpected audit page: %+v", auditPage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/audit-logs?start_time=bad-time", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/cheat-detection", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var cheat opsqry.CheatDetectionResp
	decodeFullRouterData(t, resp, &cheat)
	if cheat.Summary.SubmitBurstUsers < 1 || cheat.Summary.SharedIPGroups < 1 {
		t.Fatalf("unexpected cheat detection response: %+v", cheat)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "全员通知",
		"content": "full-router matrix admin publish",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publishResult opshttp.AdminNotificationPublishResp
	decodeFullRouterData(t, resp, &publishResult)
	if publishResult.BatchID <= 0 || publishResult.RecipientCount < 4 {
		t.Fatalf("unexpected publish result: %+v", publishResult)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "teacher forbidden",
		"content": "teacher should not publish",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "invalid audience",
		"content": "missing roles",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "role"},
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var notificationPage map[string]any
	decodeFullRouterData(t, resp, &notificationPage)
	if int(notificationPage["total"].(float64)) < 2 {
		t.Fatalf("unexpected notifications page: %+v", notificationPage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/notifications/%d/read", env.notification.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/notifications/%d/read", env.notification.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	server := httptest.NewServer(env.router)
	defer server.Close()

	ticketResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/auth/ws-ticket", nil, studentHeaders)
	assertFullRouterStatus(t, ticketResp, http.StatusOK)

	var wsTicket map[string]any
	decodeFullRouterData(t, ticketResp, &wsTicket)
	ticket, _ := wsTicket["ticket"].(string)
	if ticket == "" {
		t.Fatalf("expected websocket ticket")
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/notifications?ticket=" + ticket
	wsConfig, err := xws.NewConfig(wsURL, server.URL)
	if err != nil {
		t.Fatalf("new websocket config: %v", err)
	}
	conn, err := xws.DialConfig(wsConfig)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	message := receiveFullRouterWSMessageByType(t, conn, "system.connected")
	if message.Type != "system.connected" {
		t.Fatalf("unexpected websocket message: %+v", message)
	}

	reusedConfig, _ := xws.NewConfig(wsURL, server.URL)
	if _, err := xws.DialConfig(reusedConfig); err == nil {
		t.Fatal("expected consumed websocket ticket to be rejected")
	}
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
