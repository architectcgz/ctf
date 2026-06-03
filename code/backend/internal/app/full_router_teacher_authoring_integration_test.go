package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	"ctf-platform/internal/shared/taxonomy"
)

func TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))

	createPayload := func(title string) map[string]any {
		return map[string]any{
			"title":       title,
			"description": "ownership test challenge",
			"category":    taxonomy.DimensionWeb,
			"difficulty":  taxonomy.DifficultyEasy,
			"points":      100,
			"image_id":    env.image.ID,
		}
	}

	adminCreateResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", createPayload("admin-owned"), adminHeaders)
	assertFullRouterStatus(t, adminCreateResp, http.StatusOK)
	var adminChallenge challengehttp.ChallengeResp
	decodeFullRouterData(t, adminCreateResp, &adminChallenge)

	teacherCreateResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", createPayload("teacher-owned"), teacherHeaders)
	assertFullRouterStatus(t, teacherCreateResp, http.StatusOK)
	var teacherChallenge challengehttp.ChallengeResp
	decodeFullRouterData(t, teacherCreateResp, &teacherChallenge)

	listResp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/challenges?page=1&page_size=50", nil, teacherHeaders)
	assertFullRouterStatus(t, listResp, http.StatusOK)
	var listResult struct {
		List []challengehttp.ChallengeResp `json:"list"`
	}
	decodeFullRouterData(t, listResp, &listResult)

	foundTeacherOwned := false
	foundAdminOwned := false
	for _, item := range listResult.List {
		if item.ID == teacherChallenge.ID {
			foundTeacherOwned = true
		}
		if item.ID == adminChallenge.ID {
			foundAdminOwned = true
		}
	}
	if !foundTeacherOwned {
		t.Fatalf("teacher should see own challenge %d in list, got %+v", teacherChallenge.ID, listResult.List)
	}
	if !foundAdminOwned {
		t.Fatalf("teacher should see admin challenge %d in list, got %+v", adminChallenge.ID, listResult.List)
	}

	resp := performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID), map[string]any{
		"template_id": env.template.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID), map[string]any{
		"title":      "admin writeup",
		"content":    "admin writeup content",
		"visibility": challengeentity.WriteupVisibilityPublic,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag":        "flag{ownership-check}",
		"flag_prefix": "flag",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

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
		ChallengeID:      adminChallenge.ID,
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
		Where("id = ?", adminChallenge.ID).
		Update("status", challengecontracts.ChallengeStatusArchived).Error; err != nil {
		t.Fatalf("archive admin challenge: %v", err)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests", adminChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusAccepted)

	var publishJob challengehttp.ChallengePublishCheckJobResp
	decodeFullRouterData(t, resp, &publishJob)
	if publishJob.ChallengeID != adminChallenge.ID {
		t.Fatalf("unexpected publish request payload: %+v", publishJob)
	}

	readChecks := []struct {
		name    string
		method  string
		path    string
		payload any
		assert  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "get archived detail",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertFullRouterStatus(t, resp, http.StatusOK)
				var detail challengehttp.ChallengeResp
				decodeFullRouterData(t, resp, &detail)
				if detail.ID != adminChallenge.ID || detail.Status != challengecontracts.ChallengeStatusArchived {
					t.Fatalf("unexpected archived challenge detail: %+v", detail)
				}
			},
		},
		{
			name:   "get admin writeup",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertFullRouterStatus(t, resp, http.StatusOK)
				var writeup challengecontracts.AdminChallengeWriteupResp
				decodeFullRouterData(t, resp, &writeup)
				if writeup.Title != "admin writeup" || writeup.Visibility != challengeentity.WriteupVisibilityPublic {
					t.Fatalf("unexpected admin writeup: %+v", writeup)
				}
			},
		},
		{
			name:   "get admin flag",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertFullRouterStatus(t, resp, http.StatusOK)
				var flagResp challengehttp.FlagResp
				decodeFullRouterData(t, resp, &flagResp)
				if !flagResp.Configured || flagResp.FlagType != challengecontracts.FlagTypeStatic {
					t.Fatalf("unexpected flag config: %+v", flagResp)
				}
			},
		},
		{
			name:   "get admin topology",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertFullRouterStatus(t, resp, http.StatusOK)
				var topology challengehttp.ChallengeTopologyResp
				decodeFullRouterData(t, resp, &topology)
				if topology.TemplateID == nil || *topology.TemplateID != env.template.ID {
					t.Fatalf("unexpected topology template binding: %+v", topology)
				}
			},
		},
		{
			name:   "get latest publish request",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests/latest", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertFullRouterStatus(t, resp, http.StatusOK)
				var latest challengehttp.ChallengePublishCheckJobResp
				decodeFullRouterData(t, resp, &latest)
				if latest.ChallengeID != adminChallenge.ID || latest.Status != "queued" {
					t.Fatalf("unexpected latest publish request: %+v", latest)
				}
			},
		},
		{
			name:   "download package export",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/package-export/download?revision_id=%d", adminChallenge.ID, packageRevision.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertFullRouterStatus(t, resp, http.StatusOK)
			},
		},
	}
	for _, tc := range readChecks {
		resp := performFullRouterRequest(t, env.router, tc.method, tc.path, tc.payload, teacherHeaders)
		tc.assert(t, resp)
	}

	for _, tc := range []struct {
		name    string
		method  string
		path    string
		payload any
	}{
		{name: "update challenge", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d", adminChallenge.ID), payload: map[string]any{"title": "forbidden-update"}},
		{name: "configure flag", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID), payload: map[string]any{
			"flag_type":   challengecontracts.FlagTypeStatic,
			"flag":        "flag{ownership-check}",
			"flag_prefix": "flag",
		}},
		{name: "upsert writeup", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID), payload: map[string]any{
			"title":      "forbidden writeup",
			"content":    "forbidden content",
			"visibility": challengeentity.WriteupVisibilityPublic,
		}},
		{name: "save topology", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID), payload: map[string]any{
			"entry_node_key": "web",
			"nodes": []map[string]any{
				{
					"key":          "web",
					"name":         "web",
					"image_id":     env.image.ID,
					"service_port": 80,
					"inject_flag":  true,
					"tier":         challengecontracts.TopologyTierPublic,
				},
			},
		}},
		{name: "self check", method: http.MethodPost, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/self-check", adminChallenge.ID)},
	} {
		resp := performFullRouterRequest(t, env.router, tc.method, tc.path, tc.payload, teacherHeaders)
		if resp.Code != http.StatusForbidden {
			t.Fatalf("%s should be forbidden, got status=%d body=%s", tc.name, resp.Code, resp.Body.String())
		}
	}

	ownDetailResp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", teacherChallenge.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, ownDetailResp, http.StatusOK)

	var ownDetail challengehttp.ChallengeResp
	decodeFullRouterData(t, ownDetailResp, &ownDetail)
	if ownDetail.Status != challengecontracts.ChallengeStatusDraft {
		t.Fatalf("expected own draft challenge to stay readable, got %+v", ownDetail)
	}
}

func TestFullRouter_CreateChallengeStoresCreator(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "creator-marker",
		"description": "creator marker challenge",
		"category":    taxonomy.DimensionWeb,
		"difficulty":  taxonomy.DifficultyEasy,
		"points":      100,
		"image_id":    env.image.ID,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var challengeData map[string]any
	decodeFullRouterData(t, resp, &challengeData)
	if _, ok := challengeData["created_by"]; !ok {
		t.Fatalf("expected challenge response to include created_by, got %+v", challengeData)
	}

	challengeIDFloat, ok := challengeData["id"].(float64)
	if !ok {
		t.Fatalf("expected numeric challenge id, got %+v", challengeData["id"])
	}

	var createdBy sql.NullInt64
	if err := env.db.Raw("SELECT created_by FROM challenges WHERE id = ?", int64(challengeIDFloat)).Scan(&createdBy).Error; err != nil {
		t.Fatalf("query challenge created_by: %v", err)
	}
	if !createdBy.Valid || createdBy.Int64 != env.teacher.ID {
		t.Fatalf("unexpected created_by=%+v, want %d", createdBy, env.teacher.ID)
	}
}

func TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	resp := performFullRouterRequest(
		t,
		env.router,
		http.MethodPost,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/self-check", env.challenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var result challengehttp.ChallengeSelfCheckResp
	decodeFullRouterData(t, resp, &result)
	if result.ChallengeID != env.challenge.ID {
		t.Fatalf("expected challenge_id=%d, got %d", env.challenge.ID, result.ChallengeID)
	}
	if !result.Precheck.Passed {
		t.Fatalf("expected precheck passed, got %+v", result.Precheck)
	}
	if !result.Runtime.Passed {
		t.Fatalf("expected runtime passed, got %+v", result.Runtime)
	}
	if result.Runtime.AccessURL == "" {
		t.Fatalf("expected runtime access url, got empty")
	}
	if len(result.Runtime.Steps) == 0 {
		t.Fatalf("expected runtime steps, got empty")
	}
}
