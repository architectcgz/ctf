package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
)

type Envelope struct {
	Type      string    `json:"type"`
	Payload   any       `json:"payload,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type RetryAdvice struct {
	Strategy       string `json:"strategy"`
	InitialDelayMS int64  `json:"initial_delay_ms"`
	MaxDelayMS     int64  `json:"max_delay_ms"`
}

type connectedPayload struct {
	UserID                   int64       `json:"user_id"`
	HeartbeatIntervalSeconds int64       `json:"heartbeat_interval_seconds"`
	Retry                    RetryAdvice `json:"retry"`
}

type Manager struct {
	mu                sync.RWMutex
	clients           map[int64]map[string]*client
	channels          map[string]map[string]*client
	heartbeatInterval time.Duration
	readTimeout       time.Duration
	retryInitialDelay time.Duration
	retryMaxDelay     time.Duration
	logger            *zap.Logger
}

type client struct {
	id       string
	user     authctx.CurrentUser
	conn     *xws.Conn
	send     chan Envelope
	stop     chan struct{}
	channels map[string]struct{}
	closeMu  sync.Once
	logger   *zap.Logger
}

func NewManager(cfg config.WebSocketConfig, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Manager{
		clients:           make(map[int64]map[string]*client),
		channels:          make(map[string]map[string]*client),
		heartbeatInterval: cfg.HeartbeatInterval,
		readTimeout:       cfg.ReadTimeout,
		retryInitialDelay: cfg.RetryInitialDelay,
		retryMaxDelay:     cfg.RetryMaxDelay,
		logger:            logger,
	}
}

func (m *Manager) Serve(user authctx.CurrentUser, conn *xws.Conn) {
	m.ServeChannels(user, conn)
}

func (m *Manager) ServeChannels(user authctx.CurrentUser, conn *xws.Conn, channels ...string) {
	clientID := uuid.NewString()
	channelSet := make(map[string]struct{}, len(channels))
	for _, channel := range channels {
		if channel == "" {
			continue
		}
		channelSet[channel] = struct{}{}
	}

	connClient := &client{
		id:       clientID,
		user:     user,
		conn:     conn,
		send:     make(chan Envelope, 16),
		stop:     make(chan struct{}),
		channels: channelSet,
		logger:   m.logger.With(zap.Int64("user_id", user.UserID), zap.String("client_id", clientID)),
	}

	m.register(connClient)
	defer m.unregister(connClient)

	connClient.enqueue(Envelope{
		Type: "system.connected",
		Payload: connectedPayload{
			UserID:                   user.UserID,
			HeartbeatIntervalSeconds: int64(m.heartbeatInterval.Seconds()),
			Retry: RetryAdvice{
				Strategy:       "exponential_backoff",
				InitialDelayMS: m.retryInitialDelay.Milliseconds(),
				MaxDelayMS:     m.retryMaxDelay.Milliseconds(),
			},
		},
		Timestamp: time.Now().UTC(),
	})

	errCh := make(chan error, 2)
	go func() { errCh <- m.readLoop(connClient) }()
	go func() { errCh <- m.writeLoop(connClient) }()

	if err := <-errCh; err != nil {
		m.logger.Debug("websocket_connection_closed", zap.Int64("user_id", user.UserID), zap.Error(err))
	}
	connClient.close()
	<-errCh
}

func (m *Manager) Broadcast(message Envelope) int {
	m.mu.RLock()
	clients := make([]*client, 0)
	for _, group := range m.clients {
		for _, item := range group {
			clients = append(clients, item)
		}
	}
	m.mu.RUnlock()

	sent := 0
	for _, item := range clients {
		if item.enqueue(message) {
			sent++
		}
	}
	return sent
}

func (m *Manager) SendToUser(userID int64, message Envelope) int {
	m.mu.RLock()
	group := m.clients[userID]
	clients := make([]*client, 0, len(group))
	for _, item := range group {
		clients = append(clients, item)
	}
	m.mu.RUnlock()

	sent := 0
	for _, item := range clients {
		if item.enqueue(message) {
			sent++
		}
	}
	return sent
}

func (m *Manager) SendToChannel(channel string, message Envelope) int {
	m.mu.RLock()
	group := m.channels[channel]
	clients := make([]*client, 0, len(group))
	for _, item := range group {
		clients = append(clients, item)
	}
	m.mu.RUnlock()

	sent := 0
	for _, item := range clients {
		if item.enqueue(message) {
			sent++
		}
	}
	return sent
}

func (m *Manager) register(connClient *client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[connClient.user.UserID]; !ok {
		m.clients[connClient.user.UserID] = make(map[string]*client)
	}
	m.clients[connClient.user.UserID][connClient.id] = connClient
	for channel := range connClient.channels {
		if _, ok := m.channels[channel]; !ok {
			m.channels[channel] = make(map[string]*client)
		}
		m.channels[channel][connClient.id] = connClient
	}
}

func (m *Manager) unregister(connClient *client) {
	connClient.close()

	m.mu.Lock()
	defer m.mu.Unlock()

	group, ok := m.clients[connClient.user.UserID]
	if !ok {
		return
	}
	delete(group, connClient.id)
	if len(group) == 0 {
		delete(m.clients, connClient.user.UserID)
	}
	for channel := range connClient.channels {
		channelGroup, ok := m.channels[channel]
		if !ok {
			continue
		}
		delete(channelGroup, connClient.id)
		if len(channelGroup) == 0 {
			delete(m.channels, channel)
		}
	}
}

func (m *Manager) readLoop(connClient *client) error {
	for {
		select {
		case <-connClient.stop:
			return nil
		default:
		}

		if err := connClient.conn.SetReadDeadline(time.Now().Add(m.readTimeout)); err != nil {
			return err
		}

		var payload []byte
		if err := xws.Message.Receive(connClient.conn, &payload); err != nil {
			return err
		}
	}
}

func (m *Manager) writeLoop(connClient *client) error {
	ticker := time.NewTicker(m.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-connClient.stop:
			return nil
		case message := <-connClient.send:
			if err := connClient.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return err
			}
			if err := xws.JSON.Send(connClient.conn, message); err != nil {
				return err
			}
		case <-ticker.C:
			if err := connClient.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return err
			}
			if err := writePing(connClient.conn, []byte(fmt.Sprintf("%d", time.Now().UnixNano()))); err != nil {
				return err
			}
		}
	}
}

func (c *client) enqueue(message Envelope) bool {
	select {
	case <-c.stop:
		return false
	default:
	}

	select {
	case c.send <- message:
		return true
	case <-c.stop:
		return false
	default:
		c.logger.Warn("websocket_send_queue_full", zap.Int64("user_id", c.user.UserID), zap.String("type", message.Type))
		return false
	}
}

func (c *client) close() {
	c.closeMu.Do(func() {
		close(c.stop)
		_ = c.conn.Close()
	})
}

func writePing(conn *xws.Conn, payload []byte) error {
	writer, err := conn.NewFrameWriter(xws.PingFrame)
	if err != nil {
		return err
	}
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return err
	}
	return writer.Close()
}
