package commands

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	"ctf-platform/pkg/errcode"
)

func newContestChallengeCommandService(t *testing.T) (*ChallengeService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()

	return newContestChallengeCommandServiceWithRedis(t, nil)
}

func newContestChallengeCommandServiceWithRedis(t *testing.T, redisClient *redis.Client) (*ChallengeService, *challengeinfra.Repository, *contestinfra.Repository, *contestinfra.ChallengeRepository, *contestinfra.AWDRepository) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	awdRepo := contestinfra.NewAWDRepository(db)
	return NewChallengeService(
			contestinfra.NewChallengeRepository(db),
			challengeinfra.NewRepository(db),
			contestinfra.NewRepository(db),
			awdRepo,
			redisClient,
		),
		challengeinfra.NewRepository(db),
		contestinfra.NewRepository(db),
		contestinfra.NewChallengeRepository(db),
		awdRepo
}

func TestChallengeServiceAddChallengeToAWDContestPersistsAWDServiceConfig(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, awdRepo := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        501,
		Title:     "awd-config",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9001,
		Title:      "awd-web",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	resp, err := service.AddChallengeToContest(context.Background(), contest.ID, &dto.AddContestChallengeReq{
		ChallengeID:      9001,
		Points:           120,
		AWDCheckerType:   model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig: map[string]any{"get_flag": map[string]any{"path": "/internal/flag"}},
		AWDSLAScore:      15,
		AWDDefenseScore:  25,
	})
	if err != nil {
		t.Fatalf("AddChallengeToContest() error = %v", err)
	}
	if resp.AWDCheckerType != model.AWDCheckerTypeHTTPStandard || resp.AWDSLAScore != 15 || resp.AWDDefenseScore != 25 {
		t.Fatalf("unexpected awd challenge response: %+v", resp)
	}
	if path := resp.AWDCheckerConfig["get_flag"].(map[string]any)["path"]; path != "/internal/flag" {
		t.Fatalf("unexpected checker config: %+v", resp.AWDCheckerConfig)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	if items[0].AWDCheckerType != model.AWDCheckerTypeHTTPStandard || items[0].AWDSLAScore != 15 || items[0].AWDDefenseScore != 25 {
		t.Fatalf("unexpected stored contest challenge: %+v", items[0])
	}

	serviceAssociation, err := awdRepo.FindContestAWDServiceByContestAndChallenge(context.Background(), contest.ID, 9001)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndChallenge() error = %v", err)
	}
	if serviceAssociation.DisplayName != "awd-web" {
		t.Fatalf("unexpected contest awd service display name: %s", serviceAssociation.DisplayName)
	}
}

func TestChallengeServiceRejectsAWDServiceConfigOnNonAWDContest(t *testing.T) {
	service, challengeRepo, contestRepo, _, _ := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        502,
		Title:     "jeopardy",
		Mode:      model.ContestModeJeopardy,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9002,
		Title:      "web",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	_, err := service.AddChallengeToContest(context.Background(), contest.ID, &dto.AddContestChallengeReq{
		ChallengeID:     9002,
		AWDCheckerType:  model.AWDCheckerTypeHTTPStandard,
		AWDSLAScore:     10,
		AWDDefenseScore: 20,
	})
	if err == nil {
		t.Fatal("expected AddChallengeToContest() to reject awd config on non-awd contest")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

func TestChallengeServiceUpdateChallengePersistsAWDServiceConfig(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, awdRepo := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        503,
		Title:     "awd-update",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9003,
		Title:      "awd-update-challenge",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := challengeRelationRepo.AddChallenge(context.Background(), &model.ContestChallenge{
		ContestID:   contest.ID,
		ChallengeID: 9003,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("add challenge: %v", err)
	}

	err := service.UpdateChallenge(context.Background(), contest.ID, 9003, &dto.UpdateContestChallengeReq{
		AWDCheckerType:   stringPtr(string(model.AWDCheckerTypeHTTPStandard)),
		AWDCheckerConfig: map[string]any{"havoc": map[string]any{"path": "/healthz"}},
		AWDSLAScore:      intPtr(18),
		AWDDefenseScore:  intPtr(28),
	})
	if err != nil {
		t.Fatalf("UpdateChallenge() error = %v", err)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	if items[0].AWDCheckerType != model.AWDCheckerTypeHTTPStandard || items[0].AWDSLAScore != 18 || items[0].AWDDefenseScore != 28 {
		t.Fatalf("unexpected updated contest challenge: %+v", items[0])
	}

	serviceAssociation, err := awdRepo.FindContestAWDServiceByContestAndChallenge(context.Background(), contest.ID, 9003)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndChallenge() error = %v", err)
	}
	scoreConfig := contestdomain.ParseAWDCheckerConfig(serviceAssociation.ScoreConfig)
	if got := int(scoreConfig["awd_sla_score"].(float64)); got != 18 {
		t.Fatalf("unexpected awd service sla score: %+v", scoreConfig)
	}
}

func TestChallengeServiceAddChallengeConsumesCheckerPreviewToken(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service, challengeRepo, contestRepo, challengeRelationRepo, _ := newContestChallengeCommandServiceWithRedis(t, redisClient)

	now := time.Now()
	contest := &model.Contest{
		ID:        504,
		Title:     "awd-preview-token",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9004,
		Title:      "token-preview",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	previewServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/flag":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				_, _ = w.Write([]byte("flag{preview}"))
				return
			}
			http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(previewServer.Close)

	previewDB := contesttestsupport.SetupAWDTestDB(t)
	if err := contestinfra.NewRepository(previewDB).Create(context.Background(), &model.Contest{
		ID:        contest.ID,
		Title:     contest.Title,
		Mode:      contest.Mode,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create preview contest: %v", err)
	}
	if err := challengeinfra.NewRepository(previewDB).Create(&model.Challenge{
		ID:         9004,
		Title:      "token-preview",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create preview challenge: %v", err)
	}

	previewService := NewAWDService(
		contestinfra.NewAWDRepository(previewDB),
		contestinfra.NewRepository(previewDB),
		redisClient,
		"",
		config.ContestAWDConfig{
			CheckerTimeout:    time.Second,
			CheckerHealthPath: "/healthz",
		},
		zap.NewNop(),
		contestjobs.NewAWDRoundUpdater(
			contestinfra.NewAWDRepository(previewDB),
			redisClient,
			config.ContestAWDConfig{
				CheckerTimeout:    time.Second,
				CheckerHealthPath: "/healthz",
			},
			"",
			nil,
			zap.NewNop(),
		),
	)

	method := reflect.ValueOf(previewService).MethodByName("PreviewChecker")
	if !method.IsValid() {
		t.Fatalf("PreviewChecker method not implemented")
	}
	reqValue := reflect.New(method.Type().In(2).Elem())
	setAnyField(t, reqValue.Elem(), "ChallengeID", int64(9004))
	setAnyField(t, reqValue.Elem(), "CheckerType", string(model.AWDCheckerTypeHTTPStandard))
	setAnyField(t, reqValue.Elem(), "CheckerConfig", map[string]any{
		"put_flag": map[string]any{
			"method":          "PUT",
			"path":            "/api/flag",
			"expected_status": http.StatusCreated,
			"body_template":   "{{FLAG}}",
		},
		"get_flag": map[string]any{
			"method":             "GET",
			"path":               "/api/flag",
			"expected_status":    http.StatusOK,
			"expected_substring": "{{FLAG}}",
		},
		"havoc": map[string]any{
			"method":          "GET",
			"path":            "/healthz",
			"expected_status": http.StatusOK,
		},
	})
	setAnyField(t, reqValue.Elem(), "AccessURL", previewServer.URL)

	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(int64(504)),
		reqValue,
	})
	if errValue := results[1].Interface(); errValue != nil {
		t.Fatalf("PreviewChecker() error = %v", errValue)
	}
	previewResp := results[0]
	if previewResp.IsNil() {
		t.Fatal("expected preview response")
	}
	previewTokenField := previewResp.Elem().FieldByName("PreviewToken")
	if !previewTokenField.IsValid() || previewTokenField.String() == "" {
		t.Fatal("expected preview token")
	}

	req := &dto.AddContestChallengeReq{
		ChallengeID:    9004,
		Points:         130,
		AWDCheckerType: model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig: map[string]any{
			"put_flag": map[string]any{
				"method":          "PUT",
				"path":            "/api/flag",
				"expected_status": http.StatusCreated,
				"body_template":   "{{FLAG}}",
			},
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
			"havoc": map[string]any{
				"method":          "GET",
				"path":            "/healthz",
				"expected_status": http.StatusOK,
			},
		},
	}
	setAnyField(t, reflect.ValueOf(req).Elem(), "AWDCheckerPreviewToken", previewTokenField.String())

	resp, err := service.AddChallengeToContest(context.Background(), contest.ID, req)
	if err != nil {
		t.Fatalf("AddChallengeToContest() error = %v", err)
	}
	respValue := reflect.ValueOf(resp).Elem()
	stateField := respValue.FieldByName("AWDCheckerValidationState")
	if !stateField.IsValid() || stateField.String() != "passed" {
		t.Fatalf("expected passed validation state, got %#v", stateField)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	itemValue := reflect.ValueOf(items[0]).Elem()
	if state := itemValue.FieldByName("AWDCheckerValidationState"); !state.IsValid() || state.String() != "passed" {
		t.Fatalf("expected persisted passed validation state, got %#v", state)
	}
	if previewAt := itemValue.FieldByName("AWDCheckerLastPreviewAt"); !previewAt.IsValid() || previewAt.IsNil() {
		t.Fatal("expected persisted preview time")
	}
	if previewResult := itemValue.FieldByName("AWDCheckerLastPreviewResult"); !previewResult.IsValid() || previewResult.String() == "" {
		t.Fatal("expected persisted preview result")
	}
}

func TestChallengeServiceUpdateChallengeMarksValidationStateStaleWhenCheckerConfigChanges(t *testing.T) {
	service, challengeRepo, contestRepo, challengeRelationRepo, _ := newContestChallengeCommandService(t)

	now := time.Now()
	contest := &model.Contest{
		ID:        505,
		Title:     "awd-stale",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := contestRepo.Create(context.Background(), contest); err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := challengeRepo.Create(&model.Challenge{
		ID:         9005,
		Title:      "stale-checker",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	existing := &model.ContestChallenge{
		ContestID:        contest.ID,
		ChallengeID:      9005,
		Points:           100,
		IsVisible:        true,
		AWDCheckerType:   model.AWDCheckerTypeHTTPStandard,
		AWDCheckerConfig: `{"get_flag":{"path":"/api/flag"}}`,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	setChallengeModelField(t, existing, "AWDCheckerValidationState", "passed")
	setChallengeModelField(t, existing, "AWDCheckerLastPreviewAt", &now)
	setChallengeModelField(t, existing, "AWDCheckerLastPreviewResult", `{"service_status":"up"}`)
	if err := challengeRelationRepo.AddChallenge(context.Background(), existing); err != nil {
		t.Fatalf("add challenge: %v", err)
	}

	err := service.UpdateChallenge(context.Background(), contest.ID, 9005, &dto.UpdateContestChallengeReq{
		AWDCheckerType:   stringPtr(string(model.AWDCheckerTypeLegacyProbe)),
		AWDCheckerConfig: map[string]any{"health_path": "/healthz"},
	})
	if err != nil {
		t.Fatalf("UpdateChallenge() error = %v", err)
	}

	items, err := challengeRelationRepo.ListChallenges(context.Background(), contest.ID, false)
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected challenge count: %d", len(items))
	}
	itemValue := reflect.ValueOf(items[0]).Elem()
	if state := itemValue.FieldByName("AWDCheckerValidationState"); !state.IsValid() || state.String() != "stale" {
		t.Fatalf("expected stale validation state, got %#v", state)
	}
}

func setChallengeModelField(t *testing.T, target *model.ContestChallenge, field string, value any) {
	t.Helper()

	item := reflect.ValueOf(target).Elem().FieldByName(field)
	if !item.IsValid() {
		t.Fatalf("field %s not found", field)
	}
	if !item.CanSet() {
		t.Fatalf("field %s cannot set", field)
	}

	next := reflect.ValueOf(value)
	if next.Type().AssignableTo(item.Type()) {
		item.Set(next)
		return
	}
	if next.Type().ConvertibleTo(item.Type()) {
		item.Set(next.Convert(item.Type()))
		return
	}
	t.Fatalf("field %s type mismatch: have %s want %s", field, next.Type(), item.Type())
}

func setAnyField(t *testing.T, target reflect.Value, field string, value any) {
	t.Helper()

	item := target.FieldByName(field)
	if !item.IsValid() {
		t.Fatalf("field %s not found", field)
	}
	if !item.CanSet() {
		t.Fatalf("field %s cannot set", field)
	}

	next := reflect.ValueOf(value)
	if !next.IsValid() {
		item.Set(reflect.Zero(item.Type()))
		return
	}
	if next.Type().AssignableTo(item.Type()) {
		item.Set(next)
		return
	}
	if next.Type().ConvertibleTo(item.Type()) {
		item.Set(next.Convert(item.Type()))
		return
	}
	t.Fatalf("field %s type mismatch: have %s want %s", field, next.Type(), item.Type())
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
