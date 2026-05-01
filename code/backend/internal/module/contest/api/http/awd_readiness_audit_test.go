package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func TestUpdateContestSkipsReadinessAuditPayloadWhenCommandFailsBeforeGate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var readinessCalls int
	handler := NewHandler(
		stubContestService{
			updateContestFunc: func(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error) {
				return nil, errcode.ErrInvalidStatusTransition
			},
		},
		stubContestQueryService{
			getContestFunc: func(ctx context.Context, id int64) (*dto.ContestResp, error) {
				return &dto.ContestResp{
					ID:     id,
					Mode:   model.ContestModeAWD,
					Status: model.ContestStatusRegistration,
				}, nil
			},
		},
		stubAWDReadinessQueryService{
			getReadinessFunc: func(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error) {
				readinessCalls++
				return testAWDReadinessSnapshot(contestID), nil
			},
		},
		nil,
		nil,
	)

	ctx, recorder := newJSONTestContext(t, http.MethodPut, "/admin/contests/42", `{"status":"running","force_override":true,"override_reason":"teacher drill"}`)
	ctx.Set("id", int64(42))

	handler.UpdateContest(ctx)

	if recorder.Code != errcode.ErrInvalidStatusTransition.HTTPStatus {
		t.Fatalf("expected status %d, got %d", errcode.ErrInvalidStatusTransition.HTTPStatus, recorder.Code)
	}
	if readinessCalls != 1 {
		t.Fatalf("expected readiness snapshot prefetched once, got %d", readinessCalls)
	}
	if payload := getAWDReadinessAuditPayloadFromContext(ctx); payload != nil {
		t.Fatalf("expected no readiness audit payload, got %+v", payload)
	}
}

func TestRunCurrentRoundChecksWritesReadinessAuditPayloadAfterGateAllowsFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAWDHandler(
		stubAWDCommandService{
			runCurrentRoundChecksFunc: func(ctx context.Context, contestID int64, req *dto.RunCurrentAWDCheckerReq) (*dto.AWDCheckerRunResp, error) {
				trace := contestcmd.AWDReadinessGateTraceFromContext(ctx)
				if trace == nil {
					t.Fatal("expected readiness gate trace in command context")
				}
				trace.RecordDecision(true)
				return nil, errcode.ErrInternal
			},
		},
		stubAWDQueryService{
			getReadinessFunc: func(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error) {
				return testAWDReadinessSnapshot(contestID), nil
			},
		},
		stubAWDServiceCommandService{},
		stubAWDServiceQueryService{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodPost, "/admin/contests/42/awd/current-round/check", `{"force_override":true,"override_reason":"teacher drill"}`)
	ctx.Set("id", int64(42))

	handler.RunCurrentRoundChecks(ctx)

	if recorder.Code != errcode.ErrInternal.HTTPStatus {
		t.Fatalf("expected status %d, got %d", errcode.ErrInternal.HTTPStatus, recorder.Code)
	}
	payload := getAWDReadinessAuditPayloadFromContext(ctx)
	if payload == nil {
		t.Fatal("expected readiness audit payload")
	}
	if payload.GateAction != contestdomain.AWDReadinessActionRunCurrentRoundCheck {
		t.Fatalf("unexpected gate action: %+v", payload)
	}
	if payload.ExecutionOutcome != "failed" {
		t.Fatalf("unexpected execution outcome: %+v", payload)
	}
	if payload.ExecutionError != errcode.ErrInternal.Message {
		t.Fatalf("unexpected execution error: %+v", payload)
	}
}

type stubContestService struct {
	createContestFunc func(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error)
	updateContestFunc func(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error)
}

func (s stubContestService) CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error) {
	if s.createContestFunc != nil {
		return s.createContestFunc(ctx, req)
	}
	return nil, nil
}

func (s stubContestService) UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error) {
	if s.updateContestFunc != nil {
		return s.updateContestFunc(ctx, id, req)
	}
	return nil, nil
}

type stubContestQueryService struct {
	getContestFunc   func(ctx context.Context, id int64) (*dto.ContestResp, error)
	listContestsFunc func(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error)
}

func (s stubContestQueryService) GetContest(ctx context.Context, id int64) (*dto.ContestResp, error) {
	if s.getContestFunc != nil {
		return s.getContestFunc(ctx, id)
	}
	return nil, nil
}

func (s stubContestQueryService) ListContests(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error) {
	if s.listContestsFunc != nil {
		return s.listContestsFunc(ctx, req)
	}
	return nil, 0, nil
}

type stubAWDReadinessQueryService struct {
	getReadinessFunc func(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error)
}

func (s stubAWDReadinessQueryService) GetReadiness(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error) {
	if s.getReadinessFunc != nil {
		return s.getReadinessFunc(ctx, contestID)
	}
	return nil, nil
}

type stubAWDCommandService struct {
	createRoundFunc           func(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error)
	runCurrentRoundChecksFunc func(ctx context.Context, contestID int64, req *dto.RunCurrentAWDCheckerReq) (*dto.AWDCheckerRunResp, error)
	runRoundChecksFunc        func(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	previewCheckerFunc        func(ctx context.Context, contestID int64, req *dto.PreviewAWDCheckerReq) (*dto.AWDCheckerPreviewResp, error)
}

func (s stubAWDCommandService) CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error) {
	if s.createRoundFunc != nil {
		return s.createRoundFunc(ctx, contestID, req)
	}
	return nil, nil
}

func (s stubAWDCommandService) RunCurrentRoundChecks(ctx context.Context, contestID int64, req *dto.RunCurrentAWDCheckerReq) (*dto.AWDCheckerRunResp, error) {
	if s.runCurrentRoundChecksFunc != nil {
		return s.runCurrentRoundChecksFunc(ctx, contestID, req)
	}
	return nil, nil
}

func (s stubAWDCommandService) RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error) {
	if s.runRoundChecksFunc != nil {
		return s.runRoundChecksFunc(ctx, contestID, roundID)
	}
	return nil, nil
}

func (s stubAWDCommandService) PreviewChecker(ctx context.Context, contestID int64, req *dto.PreviewAWDCheckerReq) (*dto.AWDCheckerPreviewResp, error) {
	if s.previewCheckerFunc != nil {
		return s.previewCheckerFunc(ctx, contestID, req)
	}
	return nil, nil
}

func (stubAWDCommandService) UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error) {
	return nil, nil
}

func (stubAWDCommandService) CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error) {
	return nil, nil
}

func (stubAWDCommandService) SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error) {
	return nil, nil
}

type stubAWDQueryService struct {
	getReadinessFunc func(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error)
}

type stubAWDServiceCommandService struct{}

func (stubAWDServiceCommandService) CreateContestAWDService(ctx context.Context, contestID int64, req *dto.CreateContestAWDServiceReq) (*dto.ContestAWDServiceResp, error) {
	return nil, nil
}

func (stubAWDServiceCommandService) UpdateContestAWDService(ctx context.Context, contestID, serviceID int64, req *dto.UpdateContestAWDServiceReq) error {
	return nil
}

func (stubAWDServiceCommandService) DeleteContestAWDService(ctx context.Context, contestID, serviceID int64) error {
	return nil
}

type stubAWDServiceQueryService struct{}

func (stubAWDServiceQueryService) ListContestAWDServices(ctx context.Context, contestID int64) ([]*dto.ContestAWDServiceResp, error) {
	return nil, nil
}

func (stubAWDQueryService) ListRounds(ctx context.Context, contestID int64) ([]contestqry.AWDRoundResult, error) {
	return nil, nil
}

func (stubAWDQueryService) ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error) {
	return nil, nil
}

func (stubAWDQueryService) ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]contestqry.AWDAttackLogResult, error) {
	return nil, nil
}

func (stubAWDQueryService) GetUserWorkspace(ctx context.Context, userID, contestID int64) (*dto.ContestAWDWorkspaceResp, error) {
	return nil, nil
}

func (stubAWDQueryService) GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error) {
	return nil, nil
}

func (stubAWDQueryService) GetTrafficSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDTrafficSummaryResp, error) {
	return nil, nil
}

func (stubAWDQueryService) ListTrafficEvents(ctx context.Context, contestID, roundID int64, req *dto.ListAWDTrafficEventsReq) (*dto.AWDTrafficEventPageResp, error) {
	return nil, nil
}

func (s stubAWDQueryService) GetReadiness(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error) {
	if s.getReadinessFunc != nil {
		return s.getReadinessFunc(ctx, contestID)
	}
	return nil, nil
}

func newJSONTestContext(t *testing.T, method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("request_id", "req-http-audit-test")
	return ctx, recorder
}

func testAWDReadinessSnapshot(contestID int64) *contestqry.AWDReadinessResult {
	lastPreviewAt := time.Unix(1710000000, 0)
	accessURL := "http://checker.example/internal"
	return &contestqry.AWDReadinessResult{
		ContestID:             contestID,
		Ready:                 false,
		BlockingCount:         1,
		GlobalBlockingReasons: []string{contestdomain.AWDReadinessGlobalReasonNoChallenges},
		Items: []contestqry.AWDReadinessItem{
			{
				AWDChallengeID:  101,
				Title:           "calc",
				CheckerType:     string(model.AWDCheckerTypeHTTPStandard),
				ValidationState: string(model.AWDCheckerValidationStateFailed),
				LastPreviewAt:   &lastPreviewAt,
				LastAccessURL:   &accessURL,
				BlockingReason:  contestdomain.AWDReadinessBlockingReasonLastPreviewFailed,
			},
		},
	}
}

func getAWDReadinessAuditPayloadFromContext(c *gin.Context) *middleware.AWDReadinessAuditPayload {
	value, exists := c.Get("awd_readiness_audit_payload")
	if !exists {
		return nil
	}
	payload, _ := value.(*middleware.AWDReadinessAuditPayload)
	return payload
}
