package composition

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type AWDDefenseSSHGateway struct {
	proxyTickets runtimeHTTPProxyTicketService
	scopeReader  runtimeports.ProxyTicketInstanceReader
	executor     runtimeContainerInteractiveExecutor
	port         int
	logger       *zap.Logger

	mu       sync.Mutex
	listener net.Listener
	done     chan struct{}
}

func NewAWDDefenseSSHGateway(
	proxyTickets runtimeHTTPProxyTicketService,
	scopeReader runtimeports.ProxyTicketInstanceReader,
	executor runtimeContainerInteractiveExecutor,
	port int,
	logger *zap.Logger,
) *AWDDefenseSSHGateway {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AWDDefenseSSHGateway{
		proxyTickets: proxyTickets,
		scopeReader:  scopeReader,
		executor:     executor,
		port:         port,
		logger:       logger,
	}
}

func (g *AWDDefenseSSHGateway) Start(ctx context.Context) error {
	if g == nil || g.proxyTickets == nil || g.scopeReader == nil || g.executor == nil || g.port <= 0 {
		return nil
	}

	g.mu.Lock()
	if g.listener != nil {
		g.mu.Unlock()
		return nil
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		g.mu.Unlock()
		return err
	}
	g.listener = listener
	g.done = make(chan struct{})
	done := g.done
	g.mu.Unlock()

	config, err := g.serverConfig(ctx)
	if err != nil {
		_ = listener.Close()
		g.mu.Lock()
		g.listener = nil
		g.done = nil
		g.mu.Unlock()
		return err
	}

	go g.serve(ctx, listener, config, done)
	g.logger.Info("awd_defense_ssh_gateway_started", zap.Int("port", g.port))
	return nil
}

func (g *AWDDefenseSSHGateway) Stop(ctx context.Context) error {
	if g == nil {
		return nil
	}

	g.mu.Lock()
	listener := g.listener
	done := g.done
	g.listener = nil
	g.done = nil
	g.mu.Unlock()

	if listener == nil {
		return nil
	}
	_ = listener.Close()
	if done == nil {
		return nil
	}

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *AWDDefenseSSHGateway) serverConfig(ctx context.Context) (*ssh.ServerConfig, error) {
	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		return nil, err
	}

	config := &ssh.ServerConfig{
		ServerVersion: "SSH-2.0-CTF-AWD-Defense",
		PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			session, err := g.authenticate(ctx, meta.User(), string(password))
			if err != nil {
				return nil, err
			}
			payload, err := json.Marshal(session)
			if err != nil {
				return nil, err
			}
			return &ssh.Permissions{
				Extensions: map[string]string{
					"awd_defense_ssh_session": string(payload),
				},
			}, nil
		},
	}
	config.AddHostKey(signer)
	return config, nil
}

func (g *AWDDefenseSSHGateway) authenticate(ctx context.Context, sshUsername, password string) (*runtimeports.AWDDefenseSSHSession, error) {
	login, err := parseAWDDefenseSSHUsername(sshUsername)
	if err != nil {
		return nil, err
	}

	claims, err := g.proxyTickets.ResolveTicket(ctx, password)
	if err != nil {
		return nil, err
	}
	if claims == nil ||
		claims.Purpose != runtimeports.ProxyTicketPurposeAWDDefenseSSH ||
		claims.Username != login.username ||
		claims.ContestID == nil || *claims.ContestID != login.contestID ||
		claims.AWDServiceID == nil || *claims.AWDServiceID != login.serviceID ||
		claims.AWDAttackerTeamID == nil ||
		claims.AWDChallengeID == nil {
		return nil, errcode.ErrProxyTicketInvalid
	}

	scope, err := g.scopeReader.FindAWDDefenseSSHScope(ctx, claims.UserID, login.contestID, login.serviceID)
	if err != nil {
		return nil, err
	}
	if scope == nil ||
		scope.InstanceID != claims.InstanceID ||
		scope.TeamID != *claims.AWDAttackerTeamID ||
		scope.AWDChallengeID != *claims.AWDChallengeID ||
		scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	return &runtimeports.AWDDefenseSSHSession{
		UserID:      claims.UserID,
		Username:    claims.Username,
		InstanceID:  scope.InstanceID,
		ContestID:   scope.ContestID,
		TeamID:      scope.TeamID,
		ServiceID:   scope.ServiceID,
		ChallengeID: scope.AWDChallengeID,
		ContainerID: scope.ContainerID,
	}, nil
}

func (g *AWDDefenseSSHGateway) serve(ctx context.Context, listener net.Listener, config *ssh.ServerConfig, done chan struct{}) {
	defer close(done)

	for {
		conn, err := listener.Accept()
		if err != nil {
			g.logger.Debug("awd_defense_ssh_accept_stopped", zap.Error(err))
			return
		}
		go g.handleConn(ctx, conn, config)
	}
}

func (g *AWDDefenseSSHGateway) handleConn(ctx context.Context, rawConn net.Conn, config *ssh.ServerConfig) {
	defer rawConn.Close()

	serverConn, channels, requests, err := ssh.NewServerConn(rawConn, config)
	if err != nil {
		g.logger.Debug("awd_defense_ssh_handshake_failed", zap.Error(err))
		return
	}
	defer serverConn.Close()
	go ssh.DiscardRequests(requests)

	session, err := sshSessionFromPermissions(serverConn.Permissions)
	if err != nil {
		g.logger.Debug("awd_defense_ssh_session_decode_failed", zap.Error(err))
		return
	}

	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			g.logger.Debug("awd_defense_ssh_channel_accept_failed", zap.Error(err))
			continue
		}
		go g.handleSessionChannel(ctx, channel, requests, session)
	}
}

func (g *AWDDefenseSSHGateway) handleSessionChannel(ctx context.Context, channel ssh.Channel, requests <-chan *ssh.Request, session *runtimeports.AWDDefenseSSHSession) {
	defer channel.Close()

	started := false
	for req := range requests {
		switch req.Type {
		case "pty-req":
			_ = req.Reply(true, nil)
		case "window-change":
			_ = req.Reply(false, nil)
		case "shell":
			if started {
				_ = req.Reply(false, nil)
				continue
			}
			started = true
			_ = req.Reply(true, nil)
			g.runContainerCommand(ctx, channel, session, []string{"/bin/sh"})
			return
		case "exec":
			if started {
				_ = req.Reply(false, nil)
				continue
			}
			command := parseSSHExecCommand(req.Payload)
			if strings.TrimSpace(command) == "" {
				_ = req.Reply(false, nil)
				continue
			}
			started = true
			_ = req.Reply(true, nil)
			g.runContainerCommand(ctx, channel, session, []string{"/bin/sh", "-lc", command})
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func (g *AWDDefenseSSHGateway) runContainerCommand(ctx context.Context, channel ssh.Channel, session *runtimeports.AWDDefenseSSHSession, command []string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	status := uint32(0)
	if err := g.executor.ExecContainerInteractive(ctx, session.ContainerID, command, channel, channel); err != nil && err != io.EOF {
		status = 1
		g.logger.Warn("awd_defense_ssh_exec_failed",
			zap.Int64("instance_id", session.InstanceID),
			zap.String("container_id", session.ContainerID),
			zap.Error(err),
		)
	}
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct {
		Status uint32
	}{Status: status}))
}

type awdDefenseSSHLogin struct {
	username  string
	contestID int64
	serviceID int64
}

func parseAWDDefenseSSHUsername(input string) (*awdDefenseSSHLogin, error) {
	parts := strings.Split(input, "+")
	if len(parts) < 3 {
		return nil, errcode.ErrProxyTicketInvalid
	}
	contestID, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil || contestID <= 0 {
		return nil, errcode.ErrProxyTicketInvalid
	}
	serviceID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil || serviceID <= 0 {
		return nil, errcode.ErrProxyTicketInvalid
	}
	username := strings.Join(parts[:len(parts)-2], "+")
	if strings.TrimSpace(username) == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}
	return &awdDefenseSSHLogin{username: username, contestID: contestID, serviceID: serviceID}, nil
}

func sshSessionFromPermissions(permissions *ssh.Permissions) (*runtimeports.AWDDefenseSSHSession, error) {
	if permissions == nil || permissions.Extensions == nil {
		return nil, errcode.ErrProxyTicketInvalid
	}
	payload := permissions.Extensions["awd_defense_ssh_session"]
	if payload == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}
	var session runtimeports.AWDDefenseSSHSession
	if err := json.Unmarshal([]byte(payload), &session); err != nil {
		return nil, err
	}
	if session.ContainerID == "" || session.InstanceID <= 0 {
		return nil, errcode.ErrProxyTicketInvalid
	}
	return &session, nil
}

func parseSSHExecCommand(payload []byte) string {
	var parsed struct {
		Command string
	}
	if err := ssh.Unmarshal(payload, &parsed); err != nil {
		return ""
	}
	return parsed.Command
}
