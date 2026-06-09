package composition

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
)

const awdDefenseSSHWorkspaceDir = "/workspace"

type AWDDefenseSSHGateway struct {
	proxyTickets runtimeHTTPProxyTicketService
	scopeReader  instanceports.ProxyTicketInstanceReader
	executor     runtimeContainerInteractiveExecutor
	hostKeyPath  string
	port         int
	logger       *zap.Logger

	mu          sync.Mutex
	listener    net.Listener
	done        chan struct{}
	runCancel   context.CancelFunc
	activeConns map[net.Conn]struct{}
	workerWG    sync.WaitGroup
}

type runtimeHTTPProxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error)
	IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*instanceports.ProxyTicketClaims, error)
	ResolveAWDTargetAccessURL(ctx context.Context, claims *instanceports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error)
	MaxAge() int
}

type runtimeContainerInteractiveExecutor interface {
	ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error
}

func NewAWDDefenseSSHGateway(
	proxyTickets runtimeHTTPProxyTicketService,
	scopeReader instanceports.ProxyTicketInstanceReader,
	executor runtimeContainerInteractiveExecutor,
	hostKeyPath string,
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
		hostKeyPath:  strings.TrimSpace(hostKeyPath),
		port:         port,
		logger:       logger,
	}
}

func (g *AWDDefenseSSHGateway) Start(ctx context.Context) error {
	if g == nil || g.proxyTickets == nil || g.scopeReader == nil || g.executor == nil || g.port <= 0 {
		return nil
	}
	if ctx == nil {
		return errors.New("awd defense ssh gateway start requires context")
	}

	g.mu.Lock()
	if g.listener != nil {
		g.mu.Unlock()
		return nil
	}
	runCtx, runCancel := context.WithCancel(ctx)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		g.mu.Unlock()
		return err
	}
	g.listener = listener
	g.done = make(chan struct{})
	g.runCancel = runCancel
	if g.activeConns == nil {
		g.activeConns = make(map[net.Conn]struct{})
	}
	done := g.done
	g.mu.Unlock()

	config, err := g.serverConfig(runCtx)
	if err != nil {
		runCancel()
		_ = listener.Close()
		g.mu.Lock()
		g.listener = nil
		g.done = nil
		g.runCancel = nil
		g.mu.Unlock()
		return err
	}

	go g.serve(runCtx, listener, config, done)
	g.logger.Info("awd_defense_ssh_gateway_started", zap.Int("port", g.port))
	return nil
}

func (g *AWDDefenseSSHGateway) Stop(ctx context.Context) error {
	if g == nil {
		return nil
	}
	if ctx == nil {
		return errors.New("awd defense ssh gateway stop requires context")
	}

	g.mu.Lock()
	listener := g.listener
	done := g.done
	runCancel := g.runCancel
	activeConns := make([]net.Conn, 0, len(g.activeConns))
	for conn := range g.activeConns {
		activeConns = append(activeConns, conn)
	}
	g.listener = nil
	g.done = nil
	g.runCancel = nil
	g.mu.Unlock()

	if listener == nil && done == nil && runCancel == nil && len(activeConns) == 0 {
		return nil
	}
	if runCancel != nil {
		runCancel()
	}
	if listener != nil {
		_ = listener.Close()
	}
	for _, conn := range activeConns {
		_ = conn.Close()
	}

	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return waitWaitGroup(ctx, &g.workerWG)
}

func (g *AWDDefenseSSHGateway) serverConfig(ctx context.Context) (*ssh.ServerConfig, error) {
	_ = ctx
	signer, err := loadOrCreateAWDDefenseSSHHostKeySigner(g.hostKeyPath)
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

func loadOrCreateAWDDefenseSSHHostKeySigner(hostKeyPath string) (ssh.Signer, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(hostKeyPath))
	if cleanPath == "" || cleanPath == "." {
		return nil, fmt.Errorf("awd defense ssh host key path is empty")
	}

	signer, err := loadAWDDefenseSSHHostKeySignerFromFile(cleanPath)
	if err == nil {
		return signer, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o700); err != nil {
		return nil, fmt.Errorf("create awd defense ssh host key dir %q: %w", filepath.Dir(cleanPath), err)
	}

	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate awd defense ssh host key: %w", err)
	}
	signer, err = ssh.NewSignerFromKey(hostKey)
	if err != nil {
		return nil, fmt.Errorf("build awd defense ssh signer: %w", err)
	}

	encoded := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(hostKey),
	})
	file, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return loadAWDDefenseSSHHostKeySignerFromFile(cleanPath)
		}
		return nil, fmt.Errorf("create awd defense ssh host key %q: %w", cleanPath, err)
	}
	written := false
	defer func() {
		if !written {
			_ = file.Close()
			_ = os.Remove(cleanPath)
		}
	}()
	if _, err := file.Write(encoded); err != nil {
		return nil, fmt.Errorf("write awd defense ssh host key %q: %w", cleanPath, err)
	}
	if err := file.Chmod(0o600); err != nil {
		return nil, fmt.Errorf("chmod awd defense ssh host key %q: %w", cleanPath, err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close awd defense ssh host key %q: %w", cleanPath, err)
	}
	written = true
	return signer, nil
}

func loadAWDDefenseSSHHostKeySignerFromFile(hostKeyPath string) (ssh.Signer, error) {
	raw, err := os.ReadFile(hostKeyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(raw)
	if err != nil {
		return nil, fmt.Errorf("parse awd defense ssh host key %q: %w", hostKeyPath, err)
	}
	return signer, nil
}

func (g *AWDDefenseSSHGateway) authenticate(ctx context.Context, sshUsername, password string) (*instanceports.AWDDefenseSSHSession, error) {
	login, err := parseAWDDefenseSSHUsername(sshUsername)
	if err != nil {
		return nil, err
	}

	claims, err := g.proxyTickets.ResolveTicket(ctx, password)
	if err != nil {
		return nil, err
	}
	if claims == nil ||
		claims.Purpose != instanceports.ProxyTicketPurposeAWDDefenseSSH ||
		claims.Username != login.username ||
		claims.ContestID == nil || *claims.ContestID != login.contestID ||
		claims.AWDServiceID == nil || *claims.AWDServiceID != login.serviceID ||
		claims.AWDAttackerTeamID == nil ||
		claims.AWDChallengeID == nil {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}

	scope, err := g.scopeReader.FindAWDDefenseSSHScope(ctx, claims.UserID, login.contestID, login.serviceID)
	if err != nil {
		return nil, err
	}
	if scope == nil ||
		scope.InstanceID != claims.InstanceID ||
		scope.TeamID != *claims.AWDAttackerTeamID ||
		scope.AWDChallengeID != *claims.AWDChallengeID ||
		scope.WorkspaceRevision != *claims.AWDWorkspaceRevision ||
		scope.ContainerID == "" {
		return nil, apperror.ErrForbidden
	}

	return &instanceports.AWDDefenseSSHSession{
		UserID:            claims.UserID,
		Username:          claims.Username,
		InstanceID:        scope.InstanceID,
		ContestID:         scope.ContestID,
		TeamID:            scope.TeamID,
		ServiceID:         scope.ServiceID,
		ChallengeID:       scope.AWDChallengeID,
		WorkspaceRevision: scope.WorkspaceRevision,
		ContainerID:       scope.ContainerID,
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
		g.trackConn(conn)
		g.workerWG.Add(1)
		go g.handleConn(ctx, conn, config)
	}
}

func (g *AWDDefenseSSHGateway) handleConn(ctx context.Context, rawConn net.Conn, config *ssh.ServerConfig) {
	defer g.workerWG.Done()
	defer g.untrackConn(rawConn)
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
		g.handleSessionChannel(ctx, channel, requests, session)
	}
}

func (g *AWDDefenseSSHGateway) handleSessionChannel(ctx context.Context, channel ssh.Channel, requests <-chan *ssh.Request, session *instanceports.AWDDefenseSSHSession) {
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
			g.runContainerCommand(ctx, channel, session, []string{"/bin/sh", "-lc", fmt.Sprintf("cd %s && exec /bin/sh", awdDefenseSSHWorkspaceDir)})
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
			g.runContainerCommand(ctx, channel, session, []string{"/bin/sh", "-lc", fmt.Sprintf("cd %s && %s", awdDefenseSSHWorkspaceDir, command)})
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func (g *AWDDefenseSSHGateway) runContainerCommand(ctx context.Context, channel ssh.Channel, session *instanceports.AWDDefenseSSHSession, command []string) {
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
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	contestID, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil || contestID <= 0 {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	serviceID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil || serviceID <= 0 {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	username := strings.Join(parts[:len(parts)-2], "+")
	if strings.TrimSpace(username) == "" {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	return &awdDefenseSSHLogin{username: username, contestID: contestID, serviceID: serviceID}, nil
}

func sshSessionFromPermissions(permissions *ssh.Permissions) (*instanceports.AWDDefenseSSHSession, error) {
	if permissions == nil || permissions.Extensions == nil {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	payload := permissions.Extensions["awd_defense_ssh_session"]
	if payload == "" {
		return nil, instancecontracts.ErrProxyTicketInvalid
	}
	var session instanceports.AWDDefenseSSHSession
	if err := json.Unmarshal([]byte(payload), &session); err != nil {
		return nil, err
	}
	if session.ContainerID == "" || session.InstanceID <= 0 || session.WorkspaceRevision <= 0 {
		return nil, instancecontracts.ErrProxyTicketInvalid
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

func (g *AWDDefenseSSHGateway) trackConn(conn net.Conn) {
	if g == nil || conn == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.activeConns == nil {
		g.activeConns = make(map[net.Conn]struct{})
	}
	g.activeConns[conn] = struct{}{}
}

func (g *AWDDefenseSSHGateway) untrackConn(conn net.Conn) {
	if g == nil || conn == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.activeConns, conn)
}

func waitWaitGroup(ctx context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
