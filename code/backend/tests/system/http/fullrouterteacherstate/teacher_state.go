package fullrouterteacherstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type TeacherAccessAndRecommendationStateMatrixDriver struct {
	Request        RequestFunc
	AdminHeaders   map[string]string
	TeacherHeaders map[string]string
	StudentHeaders map[string]string
	ClassName      string
	OtherClassName string
	StudentID      int64
	OtherStudentID int64
}

type ChallengeWriteupsUseCommunitySemanticsDriver struct {
	Request                  RequestFunc
	AdminHeaders             map[string]string
	TeacherHeaders           map[string]string
	StudentHeaders           map[string]string
	CreateChallengePayload   map[string]any
	SetChallengePublished    func(t *testing.T, challengeID int64)
	CreatePracticeSubmission func(t *testing.T, challengeID int64)
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyTeacherAccessAndRecommendationStateMatrix(t *testing.T, driver TeacherAccessAndRecommendationStateMatrixDriver) {
	t.Helper()

	resp := driver.Request(http.MethodGet, "/api/v1/teacher/classes", nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var teacherClasses struct {
		List     []teachingqueryqueries.TeacherClassItem `json:"list"`
		Total    int64                                   `json:"total"`
		Page     int                                     `json:"page"`
		PageSize int                                     `json:"page_size"`
	}
	decodeEnvelopeData(t, resp, &teacherClasses)
	if teacherClasses.Total != 1 || len(teacherClasses.List) != 1 || teacherClasses.List[0].Name != driver.ClassName {
		t.Fatalf("expected only teacher class page, got %+v", teacherClasses)
	}

	resp = driver.Request(http.MethodGet, "/api/v1/teacher/classes?page=1&page_size=1", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var adminClasses struct {
		List     []teachingqueryqueries.TeacherClassItem `json:"list"`
		Total    int64                                   `json:"total"`
		Page     int                                     `json:"page"`
		PageSize int                                     `json:"page_size"`
	}
	decodeEnvelopeData(t, resp, &adminClasses)
	if adminClasses.Page != 1 || adminClasses.PageSize != 1 || len(adminClasses.List) != 1 || adminClasses.Total < 2 {
		t.Fatalf("expected admin class pagination, got %+v", adminClasses)
	}

	resp = driver.Request(http.MethodGet, "/api/v1/teacher/students?page=1&page_size=1&sort_key=total_score&sort_order=desc", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var studentDirectory struct {
		List []struct {
			ID         int64   `json:"id"`
			Username   string  `json:"username"`
			ClassName  *string `json:"class_name"`
			TotalScore int     `json:"total_score"`
		} `json:"list"`
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
	}
	decodeEnvelopeData(t, resp, &studentDirectory)
	if studentDirectory.Page != 1 || studentDirectory.PageSize != 1 || studentDirectory.Total < 2 || len(studentDirectory.List) != 1 {
		t.Fatalf("expected paged teacher student directory, got %+v", studentDirectory)
	}
	if studentDirectory.List[0].ClassName == nil || *studentDirectory.List[0].ClassName == "" {
		t.Fatalf("expected class_name in student directory item, got %+v", studentDirectory.List[0])
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/summary", driver.ClassName), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/summary", driver.OtherClassName), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/trend", driver.ClassName), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/review", driver.ClassName), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/progress", driver.StudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var progress teachingqueryqueries.TeacherProgressResp
	decodeEnvelopeData(t, resp, &progress)
	if progress.SolvedChallenges == 0 {
		t.Fatalf("expected solved challenges in teacher progress, got %+v", progress)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/timeline", driver.StudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var timeline teachingqueryqueries.TimelineResp
	decodeEnvelopeData(t, resp, &timeline)
	if len(timeline.Events) == 0 {
		t.Fatalf("expected timeline events, got %+v", timeline)
	}
	firstTimelineEvent := timeline.Events[0]
	if firstTimelineEvent.ChallengeID == 0 || firstTimelineEvent.Timestamp.IsZero() {
		t.Fatalf("expected populated timeline event fields, got %+v", firstTimelineEvent)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/recommendations", driver.StudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var teacherRecommendations teachingqueryqueries.TeacherRecommendationResp
	decodeEnvelopeData(t, resp, &teacherRecommendations)
	if len(teacherRecommendations.Challenges) == 0 {
		t.Fatalf("expected teacher recommendations, got %+v", teacherRecommendations)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/users/%d/skill-profile", driver.StudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var skillProfile assessmenthttp.SkillProfileResp
	decodeEnvelopeData(t, resp, &skillProfile)
	if skillProfile.UserID != driver.StudentID {
		t.Fatalf("expected skill profile for student %d, got %+v", driver.StudentID, skillProfile)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/progress", driver.OtherStudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/users/%d/skill-profile", driver.OtherStudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, "/api/v1/users/me/recommendations", nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var selfRecommendations assessmenthttp.RecommendationResp
	decodeEnvelopeData(t, resp, &selfRecommendations)
	if len(selfRecommendations.Challenges) == 0 {
		t.Fatalf("expected self recommendations, got %+v", selfRecommendations)
	}
}

func VerifyChallengeWriteupsUseCommunitySemantics(t *testing.T, driver ChallengeWriteupsUseCommunitySemanticsDriver) {
	t.Helper()

	resp := driver.Request(http.MethodPost, "/api/v1/authoring/challenges", driver.CreateChallengePayload, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var createdChallenge challengehttp.ChallengeResp
	decodeEnvelopeData(t, resp, &createdChallenge)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Official Solution",
		"content":    "official content",
		"visibility": challengeentity.WriteupVisibilityPublic,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.SetChallengePublished(t, createdChallenge.ID)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var officialPayload map[string]any
	decodeEnvelopeData(t, resp, &officialPayload)
	if _, ok := officialPayload["is_recommended"]; !ok {
		t.Fatalf("expected official writeup payload to expose is_recommended, got %+v", officialPayload)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "我的草稿",
		"content":           "先记入口，再写利用链。",
		"submission_status": challengeentity.SubmissionWriteupStatusDraft,
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var draftPayload map[string]any
	decodeEnvelopeData(t, resp, &draftPayload)
	if _, ok := draftPayload["review_status"]; ok {
		t.Fatalf("expected community writeup payload to drop review_status, got %+v", draftPayload)
	}

	driver.CreatePracticeSubmission(t, createdChallenge.ID)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "我的题解",
		"content":           "1. 找入口 2. 构造 payload 3. 读取 flag",
		"submission_status": "published",
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var publishedPayload map[string]any
	decodeEnvelopeData(t, resp, &publishedPayload)
	if publishedPayload["submission_status"] != "published" {
		t.Fatalf("expected published submission status, got %+v", publishedPayload)
	}
	if _, ok := publishedPayload["published_at"]; !ok {
		t.Fatalf("expected published writeup payload to expose published_at, got %+v", publishedPayload)
	}
	submissionID := int64(publishedPayload["id"].(float64))

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup/recommend", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var recommendedOfficial challengecontracts.AdminChallengeWriteupResp
	decodeEnvelopeData(t, resp, &recommendedOfficial)
	if !recommendedOfficial.IsRecommended {
		t.Fatalf("expected official writeup to become recommended, got %+v", recommendedOfficial)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/recommend", submissionID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var recommendedCommunity challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &recommendedCommunity)
	if !recommendedCommunity.IsRecommended {
		t.Fatalf("expected community writeup to become recommended, got %+v", recommendedCommunity)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/hide", submissionID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var hiddenCommunity challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &hiddenCommunity)
	if hiddenCommunity.VisibilityStatus != challengeentity.SubmissionWriteupVisibilityHidden {
		t.Fatalf("expected hidden community writeup, got %+v", hiddenCommunity)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/solutions/community", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var hiddenCommunityList struct {
		List []map[string]any `json:"list"`
	}
	decodeEnvelopeData(t, resp, &hiddenCommunityList)
	if len(hiddenCommunityList.List) != 0 {
		t.Fatalf("expected hidden community writeup to disappear from community list, got %+v", hiddenCommunityList)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/restore", submissionID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var restoredCommunity challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &restoredCommunity)
	if restoredCommunity.VisibilityStatus != challengeentity.SubmissionWriteupVisibilityVisible {
		t.Fatalf("expected restored community writeup, got %+v", restoredCommunity)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/recommend", submissionID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/solutions/recommended", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var recommendedList struct {
		List []map[string]any `json:"list"`
	}
	decodeEnvelopeData(t, resp, &recommendedList)
	if len(recommendedList.List) != 2 {
		t.Fatalf("expected recommended solutions list, got %+v", recommendedList)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/solutions/community", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var communityList struct {
		List []map[string]any `json:"list"`
	}
	decodeEnvelopeData(t, resp, &communityList)
	if len(communityList.List) != 1 {
		t.Fatalf("expected exactly one community solution, got %+v", communityList)
	}

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/recommend", submissionID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var unrecommendedCommunity challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &unrecommendedCommunity)
	if unrecommendedCommunity.IsRecommended {
		t.Fatalf("expected community writeup recommendation to be cleared, got %+v", unrecommendedCommunity)
	}

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup/recommend", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var unrecommendedOfficial challengecontracts.AdminChallengeWriteupResp
	decodeEnvelopeData(t, resp, &unrecommendedOfficial)
	if unrecommendedOfficial.IsRecommended {
		t.Fatalf("expected official writeup recommendation to be cleared, got %+v", unrecommendedOfficial)
	}
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()

	if resp.Code != want {
		t.Fatalf("expected status %d, got %d body=%s", want, resp.Code, resp.Body.String())
	}
}

func decodeEnvelopeData(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	if len(body.Data) == 0 || string(body.Data) == "null" {
		t.Fatalf("expected response data, got empty body=%s", resp.Body.String())
	}
	if err := json.Unmarshal(body.Data, target); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, resp.Body.String())
	}
}
