package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	authhttp "ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	"ctf-platform/internal/validation"
	ctfws "ctf-platform/pkg/websocket"
)

type notificationIntegrationEnv struct {
	router              *gin.Engine
	db                  *gorm.DB
	cache               *redislib.Client
	tokenService        authcontracts.TokenService
	notificationService *opscmd.NotificationService
}

type notificationTestEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

type notificationTestWSTicketResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresAt string `json:"expires_at"`
}

type notificationTestPage struct {
	List     []dto.NotificationInfo `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type notificationTestWSEnvelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

func TestHTTP_NotificationsSupportTicketListReadAndWebSocketPush(t *testing.T) {
	env := newNotificationIntegrationEnv(t)
	server := httptest.NewServer(env.router)
	defer server.Close()

	user := createNotificationUser(t, env.db, "notify_user", model.RoleStudent)
	sessionCookie, err := issueNotificationSessionCookie(env.tokenService, user)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	ticketResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/ws-ticket",
		nil,
		nil,
		[]*http.Cookie{sessionCookie},
	)
	if ticketResp.Code != http.StatusOK {
		t.Fatalf("unexpected ws-ticket status: %d body=%s", ticketResp.Code, ticketResp.Body.String())
	}
	ticketBody := decodeNotificationEnvelope(t, ticketResp)
	ticketData := decodeNotificationJSON[notificationTestWSTicketResponse](t, ticketBody.Data)
	if ticketData.Ticket == "" {
		t.Fatal("expected non-empty websocket ticket")
	}
	if _, err := time.Parse(time.RFC3339, ticketData.ExpiresAt); err != nil {
		t.Fatalf("unexpected expires_at format: %v", err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/notifications?ticket=" + ticketData.Ticket
	wsConfig, err := xws.NewConfig(wsURL, server.URL)
	if err != nil {
		t.Fatalf("new websocket config: %v", err)
	}
	conn, err := xws.DialConfig(wsConfig)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	connectedMsg := receiveWSMessageByType(t, conn, "system.connected")
	if connectedMsg.Type != "system.connected" {
		t.Fatalf("expected connected message, got %s", connectedMsg.Type)
	}

	reusedConfig, _ := xws.NewConfig(wsURL, server.URL)
	if _, err := xws.DialConfig(reusedConfig); err == nil {
		t.Fatal("expected consumed ticket to be rejected on second use")
	}

	if err := env.notificationService.SendNotification(context.Background(), user.ID, &dto.NotificationReq{
		Type:    model.NotificationTypeSystem,
		Title:   "比赛开始提醒",
		Content: "Practice 模块将于 10 分钟后维护",
	}); err != nil {
		t.Fatalf("send notification: %v", err)
	}

	pushMsg := receiveWSMessageByType(t, conn, "notification.created")
	pushData := decodeNotificationJSON[dto.NotificationInfo](t, pushMsg.Payload)
	if pushData.Title != "比赛开始提醒" {
		t.Fatalf("unexpected push title: %s", pushData.Title)
	}
	if !pushData.Unread {
		t.Fatal("expected pushed notification to be unread")
	}

	listResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/notifications?page=1&page_size=10",
		nil,
		nil,
		[]*http.Cookie{sessionCookie},
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("unexpected notifications status: %d body=%s", listResp.Code, listResp.Body.String())
	}
	listBody := decodeNotificationEnvelope(t, listResp)
	listData := decodeNotificationJSON[notificationTestPage](t, listBody.Data)
	if listData.Total != 1 || len(listData.List) != 1 {
		t.Fatalf("unexpected notifications list: %+v", listData)
	}
	if !listData.List[0].Unread {
		t.Fatal("expected unread notification in list response")
	}

	markResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodPut,
		fmt.Sprintf("/api/v1/notifications/%d/read", listData.List[0].ID),
		nil,
		nil,
		[]*http.Cookie{sessionCookie},
	)
	if markResp.Code != http.StatusOK {
		t.Fatalf("unexpected mark-as-read status: %d body=%s", markResp.Code, markResp.Body.String())
	}

	readMsg := receiveWSMessageByType(t, conn, "notification.read")
	readData := decodeNotificationJSON[dto.NotificationInfo](t, readMsg.Payload)
	if readData.Unread {
		t.Fatal("expected read push to mark notification as read")
	}

	listAfterResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/notifications?page=1&page_size=10",
		nil,
		nil,
		[]*http.Cookie{sessionCookie},
	)
	listAfterBody := decodeNotificationEnvelope(t, listAfterResp)
	listAfterData := decodeNotificationJSON[notificationTestPage](t, listAfterBody.Data)
	if listAfterData.List[0].Unread {
		t.Fatal("expected notification to be read after PUT /read")
	}
}

func TestHTTP_AdminNotificationPublishRequiresAdminAndValidPayload(t *testing.T) {
	env := newNotificationIntegrationEnv(t)

	admin := createNotificationUser(t, env.db, "admin_notify", model.RoleAdmin)
	student := createNotificationUser(t, env.db, "student_notify", model.RoleStudent)
	adminSessionCookie, err := issueNotificationSessionCookie(env.tokenService, admin)
	if err != nil {
		t.Fatalf("issue admin session: %v", err)
	}
	studentSessionCookie, err := issueNotificationSessionCookie(env.tokenService, student)
	if err != nil {
		t.Fatalf("issue student session: %v", err)
	}

	publishPayload := map[string]any{
		"type":    model.NotificationTypeSystem,
		"title":   "系统公告",
		"content": "admin publish integration test",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}

	adminResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/admin/notifications",
		publishPayload,
		nil,
		[]*http.Cookie{adminSessionCookie},
	)
	if adminResp.Code != http.StatusOK {
		t.Fatalf("unexpected publish status: %d body=%s", adminResp.Code, adminResp.Body.String())
	}
	adminBody := decodeNotificationEnvelope(t, adminResp)
	result := decodeNotificationJSON[dto.AdminNotificationPublishResp](t, adminBody.Data)
	if result.BatchID <= 0 || result.RecipientCount < 2 {
		t.Fatalf("unexpected publish result: %+v", result)
	}

	forbiddenResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/admin/notifications",
		publishPayload,
		nil,
		[]*http.Cookie{studentSessionCookie},
	)
	if forbiddenResp.Code != http.StatusForbidden {
		t.Fatalf("expected student forbidden, got %d body=%s", forbiddenResp.Code, forbiddenResp.Body.String())
	}

	invalidResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/admin/notifications",
		map[string]any{
			"type":    model.NotificationTypeSystem,
			"title":   "系统公告",
			"content": "invalid",
			"audience_rules": map[string]any{
				"mode": "union",
				"rules": []map[string]any{
					{"type": "role"},
				},
			},
		},
		nil,
		[]*http.Cookie{adminSessionCookie},
	)
	if invalidResp.Code != http.StatusBadRequest {
		t.Fatalf("expected validation failed, got %d body=%s", invalidResp.Code, invalidResp.Body.String())
	}

	listResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/notifications?page=1&page_size=10",
		nil,
		nil,
		[]*http.Cookie{studentSessionCookie},
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d body=%s", listResp.Code, listResp.Body.String())
	}
	listBody := decodeNotificationEnvelope(t, listResp)
	listData := decodeNotificationJSON[notificationTestPage](t, listBody.Data)
	if listData.Total < 1 || len(listData.List) < 1 {
		t.Fatalf("expected student to receive published notification, got %+v", listData)
	}
}

func newNotificationIntegrationEnv(t *testing.T) *notificationIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	if err := validation.Register(); err != nil {
		t.Fatalf("register validator: %v", err)
	}

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(miniRedis.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	dbPath := filepath.Join(t.TempDir(), "notification-http.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.UserRole{}, &model.NotificationBatch{}, &model.Notification{}); err != nil {
		t.Fatalf("auto migrate schema: %v", err)
	}
	seedNotificationRoles(t, db)

	authCfg, wsCfg := newNotificationTestConfigs(t)
	tokenService := authinfra.NewTokenService(authCfg, wsCfg, cache)
	authRepo := identityinfra.NewRepository(db)
	authService := authcmd.NewService(authRepo, tokenService, config.RateLimitPolicyConfig{
		Enabled:      true,
		Limit:        10,
		Window:       time.Minute,
		LockDuration: 15 * time.Minute,
	}, zap.NewNop())
	casCommandService := authcmd.NewCASService(authCfg.CAS, authRepo, tokenService, zap.NewNop(), nil)
	casQueryService := authqry.NewCASService(authCfg.CAS)
	profileCommandService := identitycmd.NewProfileService(authRepo, zap.NewNop())
	profileQueryService := identityqry.NewProfileService(authRepo)
	authHandler := authhttp.NewHandler(authService, profileCommandService, profileQueryService, tokenService, casCommandService, casQueryService, authhttp.CookieConfig{
		Name:     authCfg.SessionCookieName,
		Path:     authCfg.SessionCookiePath,
		HTTPOnly: authCfg.SessionCookieHTTPOnly,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   authCfg.SessionTTL,
	}, zap.NewNop(), nil)

	wsManager := ctfws.NewManager(wsCfg, zap.NewNop())
	notificationRepo := opsinfra.NewNotificationRepository(db)
	notificationCommandService := opscmd.NewNotificationService(notificationRepo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, wsManager, zap.NewNop())
	notificationQueryService := opsqry.NewNotificationService(notificationRepo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, zap.NewNop())
	notificationHandler := NewNotificationHandler(notificationCommandService, notificationQueryService, tokenService, wsManager, zap.NewNop())

	router := gin.New()
	router.Use(middleware.RequestID())
	apiV1 := router.Group("/api/v1")
	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService, authCfg.SessionCookieName))
	protected.POST("/auth/ws-ticket", authHandler.IssueWSTicket)
	protected.GET("/notifications", notificationHandler.ListNotifications)
	protected.PUT("/notifications/:id/read", middleware.ParseInt64Param("id"), notificationHandler.MarkAsRead)
	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.POST("/notifications", notificationHandler.PublishAdminNotification)
	router.GET("/ws/notifications", notificationHandler.ServeWS)

	return &notificationIntegrationEnv{
		router:              router,
		db:                  db,
		cache:               cache,
		tokenService:        tokenService,
		notificationService: notificationCommandService,
	}
}

func newNotificationTestConfigs(t *testing.T) (config.AuthConfig, config.WebSocketConfig) {
	t.Helper()

	return config.AuthConfig{
			SessionTTL:            time.Hour,
			SessionCookieName:     "ctf_session",
			SessionCookiePath:     "/api/v1",
			SessionCookieHTTPOnly: true,
			SessionCookieSameSite: "lax",
			SessionKeyPrefix:      "ctf:test:session",
		}, config.WebSocketConfig{
			TicketTTL:         30 * time.Second,
			TicketKeyPrefix:   "ctf:test:ws:ticket",
			HeartbeatInterval: 100 * time.Millisecond,
			ReadTimeout:       time.Second,
			RetryInitialDelay: time.Second,
			RetryMaxDelay:     5 * time.Second,
		}
}

func seedNotificationRoles(t *testing.T, db *gorm.DB) {
	t.Helper()
	roles := []model.Role{
		{ID: 1, Code: model.RoleStudent, Name: "Student"},
		{ID: 2, Code: model.RoleTeacher, Name: "Teacher"},
		{ID: 3, Code: model.RoleAdmin, Name: "Admin"},
	}
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			t.Fatalf("seed role %s: %v", role.Code, err)
		}
	}
}

func createNotificationUser(t *testing.T, db *gorm.DB, username, role string) *model.User {
	t.Helper()
	user := &model.User{
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Role:     role,
		Status:   model.UserStatusActive,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("set password: %v", err)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func performNotificationJSONRequest(
	t *testing.T,
	router http.Handler,
	method,
	target string,
	payload any,
	headers map[string]string,
	cookies []*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func issueNotificationSessionCookie(tokenService authcontracts.TokenService, user *model.User) (*http.Cookie, error) {
	session, err := tokenService.CreateSession(context.Background(), user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{Name: "ctf_session", Value: session.ID}, nil
}

func decodeNotificationEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) notificationTestEnvelope {
	t.Helper()
	var envelope notificationTestEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}

func decodeNotificationJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode payload: %v payload=%s", err, string(data))
	}
	return value
}

func receiveWSMessageByType(t *testing.T, conn *xws.Conn, expectedType string) notificationTestWSEnvelope {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		t.Fatalf("set websocket deadline: %v", err)
	}

	for {
		var message notificationTestWSEnvelope
		if err := xws.JSON.Receive(conn, &message); err != nil {
			t.Fatalf("receive websocket message: %v", err)
		}
		if message.Type == expectedType {
			return message
		}
	}
}
