package commands_test

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

func setReflectedField(t *testing.T, target reflect.Value, field string, value any) {
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

func TestAWDServicePreviewCheckerRunsWithoutPersistingServices(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()

	createAWDContestFixture(t, db, 24, now)
	createAWDChallengeFixture(t, db, 2401, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	t.Cleanup(server.Close)

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/healthz",
	})

	method := reflect.ValueOf(service.commands).MethodByName("PreviewChecker")
	if !method.IsValid() {
		t.Fatalf("PreviewChecker method not implemented")
	}

	reqValue := reflect.New(method.Type().In(2))
	setReflectedField(t, reqValue.Elem(), "AWDChallengeID", int64(2401))
	setReflectedField(t, reqValue.Elem(), "CheckerType", string(contestentity.AWDCheckerTypeHTTPStandard))
	setReflectedField(t, reqValue.Elem(), "CheckerConfig", map[string]any{
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
	setReflectedField(t, reqValue.Elem(), "AccessURL", server.URL)
	setReflectedField(t, reqValue.Elem(), "PreviewFlag", "flag{preview}")

	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(int64(24)),
		reqValue.Elem(),
	})

	if len(results) != 2 {
		t.Fatalf("unexpected result count: %d", len(results))
	}
	if errValue := results[1].Interface(); errValue != nil {
		t.Fatalf("PreviewChecker() error = %v", errValue)
	}

	respValue := results[0]
	if respValue.IsNil() {
		t.Fatal("expected preview response")
	}
	resp := respValue.Elem()
	if serviceStatus := resp.FieldByName("ServiceStatus").String(); serviceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", serviceStatus)
	}
	if checkerType := resp.FieldByName("CheckerType").String(); checkerType != string(contestentity.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker type: %s", checkerType)
	}

	checkResultField := resp.FieldByName("CheckResult")
	if !checkResultField.IsValid() || checkResultField.IsNil() {
		t.Fatal("expected check result")
	}
	checkResult, ok := checkResultField.Interface().(map[string]any)
	if !ok {
		t.Fatalf("unexpected check result type: %T", checkResultField.Interface())
	}
	if source := checkResult["check_source"]; source != "checker_preview" {
		t.Fatalf("unexpected check_source: %#v", source)
	}
	if reason := checkResult["status_reason"]; reason != "preview_quorum_passed" {
		t.Fatalf("unexpected status_reason: %#v", reason)
	}
	if summary := checkResult["preview_summary"]; summary != "3/3 通过" {
		t.Fatalf("unexpected preview_summary: %#v", summary)
	}

	previewContext := resp.FieldByName("PreviewContext")
	if !previewContext.IsValid() {
		t.Fatal("expected preview context")
	}
	if accessURL := previewContext.FieldByName("AccessURL").String(); accessURL != server.URL {
		t.Fatalf("unexpected preview access_url: %s", accessURL)
	}
	if previewFlag := previewContext.FieldByName("PreviewFlag").String(); previewFlag != "flag{preview}" {
		t.Fatalf("unexpected preview flag: %s", previewFlag)
	}
	previewToken := resp.FieldByName("PreviewToken")
	if !previewToken.IsValid() || strings.TrimSpace(previewToken.String()) == "" {
		t.Fatal("expected preview_token")
	}

	var serviceCount int64
	if err := db.Model(&contestentity.AWDTeamService{}).Count(&serviceCount).Error; err != nil {
		t.Fatalf("count awd team services: %v", err)
	}
	if serviceCount != 0 {
		t.Fatalf("expected no persisted awd team services, got %d", serviceCount)
	}
}

func TestAWDServicePreviewCheckerTCPStandardTokenMakesReadinessPassed(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	const contestID int64 = 291
	const awdChallengeID int64 = 291001
	createAWDContestFixture(t, db, contestID, now)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", contestID).Update("status", contestentity.ContestStatusRegistration).Error; err != nil {
		t.Fatalf("set contest status registration: %v", err)
	}
	createAWDChallengeFixture(t, db, awdChallengeID, now)
	if err := db.Create(&challengecontracts.AWDChallenge{
		ID:             awdChallengeID,
		Name:           "TCP Length Gate",
		Slug:           "awd-tcp-length-gate",
		Category:       "pwn",
		Difficulty:     taxonomy.DifficultyMedium,
		ServiceType:    challengecontracts.AWDServiceTypeBinaryTCP,
		DeploymentMode: challengecontracts.AWDDeploymentModeSingleContainer,
		Status:         challengecontracts.AWDChallengeStatusPublished,
		CheckerType:    contestentity.AWDCheckerTypeTCPStandard,
		CheckerConfig:  `{"timeout_ms":3000,"steps":[{"send":"PING\n","expect_contains":"PONG"},{"send_template":"SET_FLAG {{FLAG}}\n","expect_contains":"OK"},{"send":"GET_FLAG\n","expect_contains":"{{FLAG}}"}]}`,
		AccessConfig:   `{"public_base_url":"tcp://preview.internal:8080","service_port":8080}`,
		RuntimeConfig:  `{"service_port":8080,"image_ref":"ctf/awd-tcp-length-gate:latest"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	accessURL, closeTCPFixture := startAWDTCPPreviewFixture(t)
	t.Cleanup(closeTCPFixture)

	checkerConfig := map[string]any{
		"timeout_ms": 3000,
		"steps": []any{
			map[string]any{"send": "PING\n", "expect_contains": "PONG"},
			map[string]any{"send_template": "SET_FLAG {{FLAG}}\n", "expect_contains": "OK"},
			map[string]any{"send": "GET_FLAG\n", "expect_contains": "{{FLAG}}"},
		},
	}
	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout: time.Second,
	})
	preview, err := service.commands.PreviewChecker(context.Background(), contestID, contestcmd.PreviewCheckerInput{
		AWDChallengeID: awdChallengeID,
		CheckerType:    string(contestentity.AWDCheckerTypeTCPStandard),
		CheckerConfig:  checkerConfig,
		AccessURL:      accessURL,
		PreviewFlag:    "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if preview.ServiceStatus != contestentity.AWDServiceStatusUp || preview.CheckerType != contestentity.AWDCheckerTypeTCPStandard {
		t.Fatalf("unexpected preview response: %+v", preview)
	}
	if strings.TrimSpace(preview.PreviewToken) == "" {
		t.Fatal("expected preview token")
	}

	challengeRepo := challengeinfra.NewRepository(db)
	contestRepo := contestinfra.NewRepository(db)
	contestChallengeRepo := contestinfra.NewChallengeRepository(db)
	awdRepo := contestinfra.NewAWDRepository(db)
	contestService := contestcmd.NewContestAWDServiceService(
		awdRepo,
		contestRepo,
		contestChallengeRepo,
		contestinfra.NewContestChallengeLookupAdapter(challengeinfra.NewContractRepository(challengeRepo)),
		challengeRepo,
		contestinfra.NewAWDCheckerPreviewTokenStore(redisClient),
	)
	created, err := contestService.CreateContestAWDService(context.Background(), contestID, contestcmd.CreateContestAWDServiceInput{
		AWDChallengeID:         awdChallengeID,
		Points:                 100,
		Order:                  1,
		IsVisible:              boolPtr(true),
		CheckerType:            strPtr(string(contestentity.AWDCheckerTypeTCPStandard)),
		CheckerConfig:          checkerConfig,
		AWDCheckerPreviewToken: strPtr(preview.PreviewToken),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), contestID, created.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.ValidationState != contestentity.AWDCheckerValidationStatePassed {
		t.Fatalf("ValidationState = %s, want passed", stored.ValidationState)
	}
	readiness, err := contestqry.NewAWDService(awdRepo, contestRepo).GetReadiness(context.Background(), contestID)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if !readiness.Ready || readiness.PassedChallenges != 1 || readiness.BlockingCount != 0 {
		t.Fatalf("unexpected readiness: %+v", readiness)
	}
}

func startAWDTCPPreviewFixture(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleAWDTCPPreviewFixtureConn(conn)
		}
	}()

	closeFn := func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("tcp preview fixture did not stop")
		}
	}
	return "tcp://" + listener.Addr().String(), closeFn
}

func handleAWDTCPPreviewFixtureConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	storedFlag := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		switch {
		case line == "PING\n":
			_, _ = conn.Write([]byte("PONG\n"))
		case strings.HasPrefix(line, "SET_FLAG "):
			storedFlag = strings.TrimSpace(strings.TrimPrefix(line, "SET_FLAG "))
			_, _ = conn.Write([]byte("OK\n"))
		case line == "GET_FLAG\n":
			_, _ = conn.Write([]byte(storedFlag + "\n"))
			return
		default:
			_, _ = conn.Write([]byte("ERR\n"))
		}
	}
}

func TestAWDServicePreviewCheckerRejectsWhenRedisUnavailable(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now()

	createAWDContestFixture(t, db, 27, now)
	createAWDChallengeFixture(t, db, 2701, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/flag":
			if r.Method != http.MethodGet {
				http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
				return
			}
			_, _ = w.Write([]byte("flag{preview}"))
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/healthz",
	})

	_, err := service.commands.PreviewChecker(context.Background(), 27, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2701,
		CheckerType:    string(contestentity.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   server.URL,
		PreviewFlag: "flag{preview}",
	})
	if err != contestcontracts.ErrAWDCheckerPreviewUnavailable {
		t.Fatalf("expected ErrAWDCheckerPreviewUnavailable, got %v", err)
	}
}

func TestAWDServicePreviewCheckerReturnsQuorumPassWhenTwoOfThreeAttemptsSucceed(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	createAWDContestFixture(t, db, 28, now)
	createAWDChallengeFixture(t, db, 2801, now)

	roundManager := &fakeAWDPreviewRoundManager{
		previewResponses: []*contestports.AWDServicePreviewResult{
			{
				ServiceStatus: contestentity.AWDServiceStatusUp,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:00Z","targets":[{"access_url":"http://preview.internal","healthy":true,"latency_ms":21}],"put_flag":{"healthy":true,"method":"PUT","path":"/api/flag"},"get_flag":{"healthy":true,"method":"GET","path":"/api/flag"}}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2801,
				},
			},
			{
				ServiceStatus: contestentity.AWDServiceStatusDown,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"http_request_failed","checked_at":"2026-04-21T11:00:01Z","error_code":"http_request_failed","error":"connection reset by peer","targets":[{"access_url":"http://preview.internal","healthy":false,"error_code":"http_request_failed","error":"connection reset by peer"}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2801,
				},
			},
			{
				ServiceStatus: contestentity.AWDServiceStatusUp,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:02Z","targets":[{"access_url":"http://preview.internal","healthy":true,"latency_ms":18}],"put_flag":{"healthy":true,"method":"PUT","path":"/api/flag"},"get_flag":{"healthy":true,"method":"GET","path":"/api/flag"},"havoc":{"healthy":true,"method":"GET","path":"/healthz"}}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2801,
				},
			},
		},
	}

	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		"",
		config.ContestAWDConfig{},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)

	resp, err := service.PreviewChecker(context.Background(), 28, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2801,
		CheckerType:    string(contestentity.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   "http://preview.internal",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if roundManager.previewCalls != 3 {
		t.Fatalf("expected 3 preview attempts, got %d", roundManager.previewCalls)
	}
	if resp.ServiceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", resp.ServiceStatus)
	}
	if resp.CheckResult["status_reason"] != "preview_quorum_passed" {
		t.Fatalf("unexpected status_reason: %#v", resp.CheckResult["status_reason"])
	}
	if resp.CheckResult["preview_pass_count"] != float64(2) {
		t.Fatalf("unexpected preview_pass_count: %#v", resp.CheckResult["preview_pass_count"])
	}
	if resp.CheckResult["preview_total_count"] != float64(3) {
		t.Fatalf("unexpected preview_total_count: %#v", resp.CheckResult["preview_total_count"])
	}
	if resp.CheckResult["preview_summary"] != "2/3 通过" {
		t.Fatalf("unexpected preview_summary: %#v", resp.CheckResult["preview_summary"])
	}
	if resp.CheckResult["check_source"] != "checker_preview" {
		t.Fatalf("unexpected check_source: %#v", resp.CheckResult["check_source"])
	}
}

func TestAWDServicePreviewCheckerEnqueuesRealtimeProgressRelayForRequester(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	createAWDContestFixture(t, db, 281, now)
	createAWDChallengeFixture(t, db, 2811, now)

	roundManager := &fakeAWDPreviewRoundManager{
		previewResponses: []*contestports.AWDServicePreviewResult{
			{
				ServiceStatus: contestentity.AWDServiceStatusUp,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:00Z","targets":[{"access_url":"http://preview.internal","healthy":true}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2811,
				},
			},
			{
				ServiceStatus: contestentity.AWDServiceStatusUp,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:01Z","targets":[{"access_url":"http://preview.internal","healthy":true}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2811,
				},
			},
			{
				ServiceStatus: contestentity.AWDServiceStatusUp,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:02Z","targets":[{"access_url":"http://preview.internal","healthy":true}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2811,
				},
			},
		},
	}

	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	outboxRepo := contestinfra.NewRealtimeOutboxRepository(db)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		"",
		config.ContestAWDConfig{},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)
	service.SetRealtimeOutbox(outboxRepo)

	ctx := contestcmd.WithAWDPreviewRequester(context.Background(), 9001)
	_, err = service.PreviewChecker(ctx, 281, contestcmd.PreviewCheckerInput{
		AWDChallengeID:   2811,
		CheckerType:      string(contestentity.AWDCheckerTypeHTTPStandard),
		PreviewRequestID: "preview-progress-1",
		AccessURL:        "http://preview.internal",
		PreviewFlag:      "flag{preview}",
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}

	expectedPhases := []string{"prepare", "attempt-1", "attempt-2", "attempt-3", "summary"}
	pendingRelays, err := outboxRepo.ListPendingRealtimeRelays(context.Background(), time.Now().Add(time.Minute), 16)
	if err != nil {
		t.Fatalf("ListPendingRealtimeRelays() error = %v", err)
	}
	if len(pendingRelays) != len(expectedPhases) {
		t.Fatalf("expected %d pending progress relays, got %d", len(expectedPhases), len(pendingRelays))
	}
	for index, expectedPhase := range expectedPhases {
		relay := pendingRelays[index].Relay
		if relay.EventName != contestcontracts.EventAWDPreviewProgress {
			t.Fatalf("relay %d unexpected event name: %s", index, relay.EventName)
		}
		if relay.Delivery != contestcontracts.RealtimeDeliveryUser {
			t.Fatalf("relay %d unexpected delivery: %s", index, relay.Delivery)
		}
		if relay.RecipientUserID == nil || *relay.RecipientUserID != 9001 {
			t.Fatalf("relay %d unexpected recipient: %+v", index, relay.RecipientUserID)
		}
		payload, ok := relay.Payload.(contestcontracts.AWDPreviewProgressRelayPayload)
		if !ok {
			t.Fatalf("relay %d unexpected payload type: %T", index, relay.Payload)
		}
		if payload.PhaseKey != expectedPhase {
			t.Fatalf("relay %d unexpected phase_key: %s", index, payload.PhaseKey)
		}
		if payload.PreviewRequestID != "preview-progress-1" {
			t.Fatalf("relay %d unexpected preview_request_id: %s", index, payload.PreviewRequestID)
		}
		if payload.Status != "running" {
			t.Fatalf("relay %d unexpected status: %s", index, payload.Status)
		}
	}
	attemptOnePayload := pendingRelays[1].Relay.Payload.(contestcontracts.AWDPreviewProgressRelayPayload)
	if attemptOnePayload.Attempt != 1 {
		t.Fatalf("attempt-1 relay missing attempt: %+v", attemptOnePayload)
	}
	attemptThreePayload := pendingRelays[3].Relay.Payload.(contestcontracts.AWDPreviewProgressRelayPayload)
	if attemptThreePayload.Attempt != 3 {
		t.Fatalf("attempt-3 relay missing attempt: %+v", attemptThreePayload)
	}
}

func TestAWDServicePreviewCheckerReturnsQuorumFailureWhenOnlyOneAttemptSucceeds(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	createAWDContestFixture(t, db, 29, now)
	createAWDChallengeFixture(t, db, 2901, now)

	roundManager := &fakeAWDPreviewRoundManager{
		previewResponses: []*contestports.AWDServicePreviewResult{
			{
				ServiceStatus: contestentity.AWDServiceStatusDown,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"http_request_failed","checked_at":"2026-04-21T11:00:00Z","error_code":"http_request_failed","error":"connection reset by peer","targets":[{"access_url":"http://preview.internal","healthy":false,"error_code":"http_request_failed"}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2901,
				},
			},
			{
				ServiceStatus: contestentity.AWDServiceStatusUp,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:01Z","targets":[{"access_url":"http://preview.internal","healthy":true,"latency_ms":20}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2901,
				},
			},
			{
				ServiceStatus: contestentity.AWDServiceStatusDown,
				CheckerType:   contestentity.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"unexpected_http_status","checked_at":"2026-04-21T11:00:02Z","error_code":"unexpected_http_status","error":"unexpected status 502","targets":[{"access_url":"http://preview.internal","healthy":false,"error_code":"unexpected_http_status"}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2901,
				},
			},
		},
	}

	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		"",
		config.ContestAWDConfig{},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)

	resp, err := service.PreviewChecker(context.Background(), 29, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2901,
		CheckerType:    string(contestentity.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   "http://preview.internal",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if roundManager.previewCalls != 3 {
		t.Fatalf("expected 3 preview attempts, got %d", roundManager.previewCalls)
	}
	if resp.ServiceStatus != contestentity.AWDServiceStatusDown {
		t.Fatalf("unexpected service status: %s", resp.ServiceStatus)
	}
	if resp.CheckResult["status_reason"] != "preview_quorum_failed" {
		t.Fatalf("unexpected status_reason: %#v", resp.CheckResult["status_reason"])
	}
	if resp.CheckResult["preview_pass_count"] != float64(1) {
		t.Fatalf("unexpected preview_pass_count: %#v", resp.CheckResult["preview_pass_count"])
	}
	if resp.CheckResult["preview_total_count"] != float64(3) {
		t.Fatalf("unexpected preview_total_count: %#v", resp.CheckResult["preview_total_count"])
	}
	if resp.CheckResult["preview_summary"] != "1/3 通过" {
		t.Fatalf("unexpected preview_summary: %#v", resp.CheckResult["preview_summary"])
	}
	if resp.CheckResult["error_code"] != "http_request_failed" {
		t.Fatalf("unexpected error_code: %#v", resp.CheckResult["error_code"])
	}
}

func TestAWDServicePreviewCheckerAcceptsServiceIDAndReturnsServiceContext(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()

	createAWDContestFixture(t, db, 25, now)
	createAWDChallengeFixture(t, db, 2501, now)
	createAWDContestChallengeFixture(t, db, 25, 2501, now)
	serviceID := defaultAWDContestServiceID(25, 2501)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/healthz",
	})

	method := reflect.ValueOf(service.commands).MethodByName("PreviewChecker")
	if !method.IsValid() {
		t.Fatalf("PreviewChecker method not implemented")
	}

	reqValue := reflect.New(method.Type().In(2))
	setReflectedField(t, reqValue.Elem(), "ServiceID", serviceID)
	setReflectedField(t, reqValue.Elem(), "CheckerType", string(contestentity.AWDCheckerTypeHTTPStandard))
	setReflectedField(t, reqValue.Elem(), "CheckerConfig", map[string]any{
		"get_flag": map[string]any{
			"method":             "GET",
			"path":               "/healthz",
			"expected_status":    http.StatusOK,
			"expected_substring": "",
		},
	})
	setReflectedField(t, reqValue.Elem(), "AccessURL", server.URL)

	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(int64(25)),
		reqValue.Elem(),
	})

	if len(results) != 2 {
		t.Fatalf("unexpected result count: %d", len(results))
	}
	if errValue := results[1].Interface(); errValue != nil {
		t.Fatalf("PreviewChecker() error = %v", errValue)
	}

	respValue := results[0]
	if respValue.IsNil() {
		t.Fatal("expected preview response")
	}
	resp := respValue.Elem()
	previewContext := resp.FieldByName("PreviewContext")
	if !previewContext.IsValid() {
		t.Fatal("expected preview context")
	}
	if gotServiceID := previewContext.FieldByName("ServiceID").Int(); gotServiceID != serviceID {
		t.Fatalf("unexpected preview service_id: %d", gotServiceID)
	}
	if gotChallengeID := previewContext.FieldByName("AWDChallengeID").Int(); gotChallengeID != 2501 {
		t.Fatalf("unexpected preview awd_challenge_id: %d", gotChallengeID)
	}
	previewToken := resp.FieldByName("PreviewToken")
	if !previewToken.IsValid() || strings.TrimSpace(previewToken.String()) == "" {
		t.Fatal("expected preview_token")
	}
}

func TestAWDServicePreviewCheckerStartsPreviewRuntimeWhenAccessURLMissing(t *testing.T) {
	db := newAWDTestDB(t)
	if err := db.AutoMigrate(&contestCommandImageRow{}, &challengecontracts.AWDChallenge{}); err != nil {
		t.Fatalf("auto migrate preview dependencies: %v", err)
	}
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	const previewSecret = "preview-secret-12345678901234567890"
	createAWDContestFixture(t, db, 26, now)
	if err := db.Create(&contestCommandImageRow{
		ID:        26001,
		Name:      "registry.example.edu/ctf/awd-preview",
		Tag:       "v1",
		Digest:    "sha256:preview-v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&challengecontracts.AWDChallenge{
		ID:             2601,
		Name:           "Preview Target",
		Slug:           "preview-target",
		Category:       "web",
		Difficulty:     taxonomy.DifficultyEasy,
		ServiceType:    challengecontracts.AWDServiceTypeWebHTTP,
		DeploymentMode: challengecontracts.AWDDeploymentModeSingleContainer,
		Status:         challengecontracts.AWDChallengeStatusPublished,
		CheckerType:    contestentity.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}}`,
		RuntimeConfig:  `{"image_id":26001,"image_ref":"registry.example.edu/ctf/awd-preview:v1","checker_token_env":"CHECKER_TOKEN"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/flag" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("flag{preview}"))
	}))
	t.Cleanup(server.Close)

	runtimeProbe := &fakeContestPreviewRuntimeProbe{
		containerAccessURL: server.URL,
		containerDetails: runtimecontracts.InstanceRuntimeDetails{
			Containers: []runtimecontracts.InstanceRuntimeContainer{{ContainerID: "preview-container"}},
		},
	}
	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		previewSecret,
		config.ContestAWDConfig{
			CheckerTimeout:    time.Second,
			CheckerHealthPath: "/healthz",
		},
		zap.NewNop(),
		newAWDCommandRoundManagerForTest(db, redisClient, config.ContestAWDConfig{
			CheckerTimeout:    time.Second,
			CheckerHealthPath: "/healthz",
		}, previewSecret, nil, zap.NewNop()),
		imageRepo,
		awdChallengeRepo,
		runtimeProbe,
	)

	resp, err := service.PreviewChecker(context.Background(), 26, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2601,
		CheckerType:    string(contestentity.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if !runtimeProbe.createContainerCalled {
		t.Fatal("expected preview runtime container startup")
	}
	if !runtimeProbe.cleanupCalled {
		t.Fatal("expected preview runtime cleanup")
	}
	if runtimeProbe.lastImageName != "registry.example.edu/ctf/awd-preview@sha256:preview-v1" {
		t.Fatalf("unexpected preview image: %s", runtimeProbe.lastImageName)
	}
	if runtimeProbe.lastEnv["FLAG"] != "flag{preview}" {
		t.Fatalf("unexpected preview FLAG env: %+v", runtimeProbe.lastEnv)
	}
	if runtimeProbe.lastEnv["CHECKER_TOKEN"] != contestdomain.BuildAWDCheckerPreviewToken(26, 0, 2601, previewSecret) {
		t.Fatalf("unexpected preview CHECKER_TOKEN env: %+v", runtimeProbe.lastEnv)
	}
	if resp.PreviewContext.AccessURL != server.URL {
		t.Fatalf("unexpected preview access url: %s", resp.PreviewContext.AccessURL)
	}
	if resp.ServiceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", resp.ServiceStatus)
	}
}

func TestAWDServicePreviewCheckerRejectsExplicitAccessURLWhenRuntimeImageUnavailable(t *testing.T) {
	db := newAWDTestDB(t)
	if err := db.AutoMigrate(&challengecontracts.AWDChallenge{}); err != nil {
		t.Fatalf("auto migrate awd challenge: %v", err)
	}
	now := time.Now()
	createAWDContestFixture(t, db, 260, now)
	if err := db.Create(&contestCommandImageRow{
		ID:        26002,
		Name:      "registry.example.edu/ctf/awd-preview",
		Tag:       "pending",
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&challengecontracts.AWDChallenge{
		ID:             2602,
		Name:           "Preview Pending Image",
		Slug:           "preview-pending-image",
		Category:       "web",
		Difficulty:     taxonomy.DifficultyEasy,
		ServiceType:    challengecontracts.AWDServiceTypeWebHTTP,
		DeploymentMode: challengecontracts.AWDDeploymentModeSingleContainer,
		Status:         challengecontracts.AWDChallengeStatusPublished,
		CheckerType:    contestentity.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}}`,
		RuntimeConfig:  `{"image_id":26002,"image_ref":"registry.example.edu/ctf/awd-preview:pending"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	roundManager := &fakeAWDPreviewRoundManager{}
	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(nil)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		nil,
		previewTokenStore,
		"",
		config.ContestAWDConfig{CheckerTimeout: time.Second},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)

	_, err := service.PreviewChecker(context.Background(), 260, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2602,
		CheckerType:    string(contestentity.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   "http://preview.internal",
		PreviewFlag: "flag{preview}",
	})
	if err == nil {
		t.Fatal("expected PreviewChecker() to reject unavailable runtime image")
	}
	if len(roundManager.previewRequests) != 0 {
		t.Fatalf("preview should be blocked before checker execution, got %+v", roundManager.previewRequests)
	}
}
