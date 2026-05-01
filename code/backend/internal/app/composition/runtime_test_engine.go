package composition

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type testRuntimeEngine struct {
	mu             sync.Mutex
	logger         *zap.Logger
	nextID         int64
	networksByID   map[string]*testRuntimeNetwork
	networksByName map[string]*testRuntimeNetwork
	containers     map[string]*testRuntimeContainer
}

type testRuntimeNetwork struct {
	id       string
	name     string
	internal bool
}

type testRuntimeContainer struct {
	id          string
	hostPort    int
	servicePort int
	networks    map[string]string
	files       map[string][]byte
	serverRef   *sharedTestRuntimeHTTPServer
	createdAt   time.Time
}

type sharedTestRuntimeHTTPServer struct {
	port     int
	listener net.Listener
	server   *http.Server
	refs     int
}

var (
	sharedTestRuntimeHTTPServersMu sync.Mutex
	sharedTestRuntimeHTTPServers   = make(map[int]*sharedTestRuntimeHTTPServer)
)

func newTestRuntimeEngine(logger *zap.Logger) *testRuntimeEngine {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &testRuntimeEngine{
		logger:         logger,
		networksByID:   make(map[string]*testRuntimeNetwork),
		networksByName: make(map[string]*testRuntimeNetwork),
		containers:     make(map[string]*testRuntimeContainer),
	}
}

func (e *testRuntimeEngine) nextIdentifier(prefix string) string {
	e.nextID++
	return fmt.Sprintf("%s-%d", prefix, e.nextID)
}

func (e *testRuntimeEngine) CreateNetwork(_ context.Context, name string, _ map[string]string, internal bool, allowExisting bool) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if existing, ok := e.networksByName[name]; ok {
		if !allowExisting {
			return "", fmt.Errorf("network %q already exists", name)
		}
		return existing.id, nil
	}

	id := e.nextIdentifier("test-net")
	network := &testRuntimeNetwork{id: id, name: name, internal: internal}
	e.networksByID[id] = network
	e.networksByName[name] = network
	return id, nil
}

func (e *testRuntimeEngine) CreateContainer(_ context.Context, cfg *model.ContainerConfig) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := e.nextIdentifier("test-ctr")
	container := &testRuntimeContainer{
		id:        id,
		networks:  make(map[string]string),
		files:     make(map[string][]byte),
		createdAt: time.Now(),
	}

	for _, hostPort := range cfg.Ports {
		value, err := strconv.Atoi(hostPort)
		if err != nil {
			return "", err
		}
		container.hostPort = value
		break
	}

	for containerPort := range cfg.Ports {
		value, err := strconv.Atoi(containerPort)
		if err != nil {
			return "", err
		}
		container.servicePort = value
		break
	}

	if strings.TrimSpace(cfg.Network) != "" {
		network, ok := e.networksByName[cfg.Network]
		if !ok {
			network = &testRuntimeNetwork{
				id:   e.nextIdentifier("test-net"),
				name: cfg.Network,
			}
			e.networksByID[network.id] = network
			e.networksByName[network.name] = network
		}
		container.networks[network.name] = e.allocateIPLocked(network.name)
	}

	e.containers[id] = container
	return id, nil
}

func (e *testRuntimeEngine) ResolveServicePort(_ context.Context, _ string, preferredPort int) (int, error) {
	if preferredPort > 0 {
		return preferredPort, nil
	}
	return 80, nil
}

func (e *testRuntimeEngine) ConnectContainerToNetwork(_ context.Context, containerID, networkName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	container, ok := e.containers[containerID]
	if !ok {
		return nil
	}
	network, ok := e.networksByName[networkName]
	if !ok {
		network = &testRuntimeNetwork{
			id:   e.nextIdentifier("test-net"),
			name: networkName,
		}
		e.networksByID[network.id] = network
		e.networksByName[network.name] = network
	}
	container.networks[network.name] = e.allocateIPLocked(network.name)
	return nil
}

func (e *testRuntimeEngine) InspectContainerNetworkIPs(_ context.Context, containerID string) (map[string]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	container, ok := e.containers[containerID]
	if !ok {
		return map[string]string{}, nil
	}
	result := make(map[string]string, len(container.networks))
	for key, value := range container.networks {
		result[key] = value
	}
	return result, nil
}

func (e *testRuntimeEngine) StartContainer(_ context.Context, containerID string) error {
	e.mu.Lock()
	container, ok := e.containers[containerID]
	if !ok || container.hostPort <= 0 || container.serverRef != nil {
		e.mu.Unlock()
		return nil
	}
	e.mu.Unlock()

	serverRef, err := acquireSharedTestRuntimeHTTPServer(container.hostPort, e.logger, containerID)
	if err != nil {
		return err
	}

	e.mu.Lock()
	if current, ok := e.containers[containerID]; ok {
		current.serverRef = serverRef
	}
	e.mu.Unlock()
	return nil
}

func (e *testRuntimeEngine) StopContainer(ctx context.Context, containerID string, _ time.Duration) error {
	e.mu.Lock()
	container, ok := e.containers[containerID]
	if !ok || container.serverRef == nil {
		e.mu.Unlock()
		return nil
	}
	serverRef := container.serverRef
	container.serverRef = nil
	e.mu.Unlock()

	return releaseSharedTestRuntimeHTTPServer(ctx, serverRef)
}

func (e *testRuntimeEngine) RemoveContainer(ctx context.Context, containerID string, _ bool) error {
	if err := e.StopContainer(ctx, containerID, time.Second); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.containers, containerID)
	return nil
}

func (e *testRuntimeEngine) RemoveNetwork(_ context.Context, networkID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	network, ok := e.networksByID[networkID]
	if !ok {
		return nil
	}
	delete(e.networksByID, networkID)
	delete(e.networksByName, network.name)
	return nil
}

func (e *testRuntimeEngine) ApplyACLRules(_ context.Context, _ []model.InstanceRuntimeACLRule) error {
	return nil
}

func (e *testRuntimeEngine) RemoveACLRules(_ context.Context, _ []model.InstanceRuntimeACLRule) error {
	return nil
}

func (e *testRuntimeEngine) ReadFileFromContainer(_ context.Context, containerID, filePath string, limit int64) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	container, ok := e.containers[containerID]
	if !ok {
		return nil, fmt.Errorf("container not found")
	}
	content, ok := container.files[filePath]
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	if limit > 0 && int64(len(content)) > limit {
		return nil, fmt.Errorf("file exceeds limit")
	}
	return append([]byte(nil), content...), nil
}

func (e *testRuntimeEngine) ListDirectoryFromContainer(_ context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	container, ok := e.containers[containerID]
	if !ok {
		return nil, fmt.Errorf("container not found")
	}
	if limit <= 0 {
		limit = 300
	}
	if dirPath == "" {
		dirPath = "."
	}
	prefix := ""
	if dirPath != "." {
		prefix = strings.TrimSuffix(dirPath, "/") + "/"
	}

	entriesByName := make(map[string]runtimeports.ContainerDirectoryEntry)
	for filePath, content := range container.files {
		if prefix != "" && !strings.HasPrefix(filePath, prefix) {
			continue
		}
		rel := strings.TrimPrefix(filePath, prefix)
		parts := strings.Split(rel, "/")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		entryType := "file"
		size := int64(len(content))
		if len(parts) > 1 {
			entryType = "dir"
			size = 0
		}
		name := parts[0]
		if existing, exists := entriesByName[name]; !exists || existing.Type != "dir" {
			entriesByName[name] = runtimeports.ContainerDirectoryEntry{Name: name, Type: entryType, Size: size}
		}
		if len(entriesByName) >= limit {
			break
		}
	}
	entries := make([]runtimeports.ContainerDirectoryEntry, 0, len(entriesByName))
	for _, entry := range entriesByName {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (e *testRuntimeEngine) WriteFileToContainer(_ context.Context, containerID, filePath string, content []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	container, ok := e.containers[containerID]
	if !ok {
		return fmt.Errorf("container not found")
	}
	if container.files == nil {
		container.files = make(map[string][]byte)
	}
	container.files[filePath] = append([]byte(nil), content...)
	return nil
}

func (e *testRuntimeEngine) ExecContainerCommand(_ context.Context, containerID string, command []string, _ []byte, _ int64) ([]byte, error) {
	e.mu.Lock()
	_, ok := e.containers[containerID]
	e.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("container not found")
	}
	return []byte(strings.Join(command, " ")), nil
}

func (e *testRuntimeEngine) ExecContainerInteractive(ctx context.Context, containerID string, _ []string, stdin io.Reader, stdout io.Writer) error {
	e.mu.Lock()
	_, ok := e.containers[containerID]
	e.mu.Unlock()
	if !ok {
		return nil
	}

	_, _ = io.WriteString(stdout, "test shell\n")
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, stdin)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (e *testRuntimeEngine) InspectImageSize(_ context.Context, _ string) (int64, error) {
	return 1, nil
}

func (e *testRuntimeEngine) RemoveImage(_ context.Context, _ string) error {
	return nil
}

func (e *testRuntimeEngine) ListManagedContainers(_ context.Context) ([]runtimeports.ManagedContainer, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]runtimeports.ManagedContainer, 0, len(e.containers))
	for _, container := range e.containers {
		result = append(result, runtimeports.ManagedContainer{
			ID:        container.id,
			Name:      container.id,
			CreatedAt: container.createdAt,
		})
	}
	return result, nil
}

func (e *testRuntimeEngine) InspectManagedContainer(_ context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	container, ok := e.containers[containerID]
	if !ok {
		return &runtimeports.ManagedContainerState{ID: containerID, Exists: false}, nil
	}
	return &runtimeports.ManagedContainerState{
		ID:      container.id,
		Exists:  true,
		Running: true,
		Status:  "running",
	}, nil
}

func (e *testRuntimeEngine) ListManagedContainerStats(_ context.Context) ([]runtimeports.ManagedContainerStat, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]runtimeports.ManagedContainerStat, 0, len(e.containers))
	for _, container := range e.containers {
		result = append(result, runtimeports.ManagedContainerStat{
			ContainerID:   container.id,
			ContainerName: container.id,
		})
	}
	return result, nil
}

func (e *testRuntimeEngine) allocateIPLocked(networkName string) string {
	segments := strings.Split(networkName, "-")
	seed := len(segments[len(segments)-1]) + len(e.containers) + len(e.networksByName)
	return fmt.Sprintf("172.30.%d.%d", seed%250+1, (e.nextID%200)+2)
}

func acquireSharedTestRuntimeHTTPServer(port int, logger *zap.Logger, containerID string) (*sharedTestRuntimeHTTPServer, error) {
	sharedTestRuntimeHTTPServersMu.Lock()
	if existing, ok := sharedTestRuntimeHTTPServers[port]; ok {
		existing.refs++
		sharedTestRuntimeHTTPServersMu.Unlock()
		return existing, nil
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		sharedTestRuntimeHTTPServersMu.Unlock()
		return nil, err
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test runtime ok"))
		}),
	}
	entry := &sharedTestRuntimeHTTPServer{
		port:     port,
		listener: listener,
		server:   server,
		refs:     1,
	}
	sharedTestRuntimeHTTPServers[port] = entry
	sharedTestRuntimeHTTPServersMu.Unlock()

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Warn("test_runtime_engine_serve_failed", zap.String("container_id", containerID), zap.Int("host_port", port), zap.Error(err))
		}
	}()
	return entry, nil
}

func releaseSharedTestRuntimeHTTPServer(ctx context.Context, entry *sharedTestRuntimeHTTPServer) error {
	if entry == nil {
		return nil
	}

	sharedTestRuntimeHTTPServersMu.Lock()
	current, ok := sharedTestRuntimeHTTPServers[entry.port]
	if !ok || current != entry {
		sharedTestRuntimeHTTPServersMu.Unlock()
		return nil
	}
	current.refs--
	if current.refs > 0 {
		sharedTestRuntimeHTTPServersMu.Unlock()
		return nil
	}
	delete(sharedTestRuntimeHTTPServers, entry.port)
	sharedTestRuntimeHTTPServersMu.Unlock()

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err := entry.server.Shutdown(shutdownCtx)
	_ = entry.listener.Close()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
