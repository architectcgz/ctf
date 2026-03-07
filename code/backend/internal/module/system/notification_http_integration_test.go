package system

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	authModule "ctf-platform/internal/module/auth"
	"ctf-platform/internal/validation"
	jwtpkg "ctf-platform/pkg/jwt"
	ctfws "ctf-platform/pkg/websocket"
)

type notificationIntegrationEnv struct {
	router              *gin.Engine
	db                  *gorm.DB
	cache               *redislib.Client
	tokenService        authModule.TokenService
	notificationService *NotificationService
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
	tokens, err := env.tokenService.IssueTokens(user.ID, user.Username, user.Role)
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	ticketResp := performNotificationJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/auth/ws-ticket",
		nil,
		map[string]string{"Authorization": "Bearer " + tokens.AccessToken},
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
		map[string]string{"Authorization": "Bearer " + tokens.AccessToken},
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
		map[string]string{"Authorization": "Bearer " + tokens.AccessToken},
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
		map[string]string{"Authorization": "Bearer " + tokens.AccessToken},
	)
	listAfterBody := decodeNotificationEnvelope(t, listAfterResp)
	listAfterData := decodeNotificationJSON[notificationTestPage](t, listAfterBody.Data)
	if listAfterData.List[0].Unread {
		t.Fatal("expected notification to be read after PUT /read")
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
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.UserRole{}, &model.Notification{}); err != nil {
		t.Fatalf("auto migrate schema: %v", err)
	}
	seedNotificationRoles(t, db)

	authCfg, wsCfg := newNotificationTestConfigs(t)
	jwtManager, err := jwtpkg.NewManager(authCfg, "ctf-platform-test")
	if err != nil {
		t.Fatalf("create jwt manager: %v", err)
	}
	tokenService := authModule.NewTokenService(authCfg, wsCfg, cache, jwtManager)
	authRepo := authModule.NewRepository(db)
	authService := authModule.NewService(authRepo, tokenService, zap.NewNop())
	authHandler := authModule.NewHandler(authService, tokenService, authModule.CookieConfig{
		Name:     authCfg.RefreshCookieName,
		Path:     authCfg.RefreshCookiePath,
		HTTPOnly: authCfg.RefreshCookieHTTPOnly,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   authCfg.RefreshTokenTTL,
	}, zap.NewNop(), nil)

	wsManager := ctfws.NewManager(wsCfg, zap.NewNop())
	notificationRepo := NewNotificationRepository(db)
	notificationService := NewNotificationService(notificationRepo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, wsManager, zap.NewNop())
	notificationHandler := NewNotificationHandler(notificationService, tokenService, wsManager, zap.NewNop())

	router := gin.New()
	router.Use(middleware.RequestID())
	apiV1 := router.Group("/api/v1")
	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService))
	protected.POST("/auth/ws-ticket", authHandler.IssueWSTicket)
	protected.GET("/notifications", notificationHandler.ListNotifications)
	protected.PUT("/notifications/:id/read", middleware.ParseInt64Param("id"), notificationHandler.MarkAsRead)
	router.GET("/ws/notifications", notificationHandler.ServeWS)

	return &notificationIntegrationEnv{
		router:              router,
		db:                  db,
		cache:               cache,
		tokenService:        tokenService,
		notificationService: notificationService,
	}
}

func newNotificationTestConfigs(t *testing.T) (config.AuthConfig, config.WebSocketConfig) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})

	keyDir := t.TempDir()
	privatePath := filepath.Join(keyDir, "private.pem")
	publicPath := filepath.Join(keyDir, "public.pem")
	if err := os.WriteFile(privatePath, privatePEM, 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	if err := os.WriteFile(publicPath, publicPEM, 0o644); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return config.AuthConfig{
			Issuer:                "ctf-platform-test",
			AccessTokenTTL:        10 * time.Minute,
			RefreshTokenTTL:       time.Hour,
			RefreshCookieName:     "refresh_token",
			RefreshCookiePath:     "/api/v1/auth",
			RefreshCookieHTTPOnly: true,
			RefreshCookieSameSite: "lax",
			PrivateKeyPath:        privatePath,
			PublicKeyPath:         publicPath,
			TokenBlacklistPrefix:  "ctf:test:blacklist",
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

func performNotificationJSONRequest(t *testing.T, router http.Handler, method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
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

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
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
