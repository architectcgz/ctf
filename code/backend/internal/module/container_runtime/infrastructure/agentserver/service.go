package agentserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/container_runtime/agentcontracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	errRuntimeExecutorUnavailable = errors.New("runtime host executor is unavailable")
	errSandboxExecutorUnavailable = errors.New("sandbox executor is unavailable")
)

type Service struct {
	hostExecutor    runtimeports.RuntimeHostExecutor
	sandboxExecutor runtimeports.SandboxExecutor
}

func NewService(hostExecutor runtimeports.RuntimeHostExecutor, sandboxExecutor runtimeports.SandboxExecutor) *Service {
	return &Service{hostExecutor: hostExecutor, sandboxExecutor: sandboxExecutor}
}

func LoadServerTLSConfig(certFile, keyFile, clientCAFile string) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load runtime agent server key pair: %w", err)
	}
	caPEM, err := os.ReadFile(clientCAFile)
	if err != nil {
		return nil, fmt.Errorf("read runtime agent server client ca: %w", err)
	}
	clientCAPool := x509.NewCertPool()
	if !clientCAPool.AppendCertsFromPEM(caPEM) {
		return nil, errors.New("append runtime agent server client ca pem failed")
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAPool,
	}, nil
}

func NewGRPCServer(tlsConfig *tls.Config, service *Service, serverCfg config.RuntimeAgentServerConfig) *grpc.Server {
	options := []grpc.ServerOption{
		grpc.ForceServerCodec(agentcontracts.JSONCodec()),
		grpc.KeepaliveEnforcementPolicy(runtimeAgentKeepaliveEnforcementPolicy(serverCfg)),
	}
	if tlsConfig != nil {
		options = append(options, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	server := grpc.NewServer(options...)
	agentcontracts.RegisterRuntimeAgentService(server, service)
	return server
}

func runtimeAgentKeepaliveEnforcementPolicy(serverCfg config.RuntimeAgentServerConfig) keepalive.EnforcementPolicy {
	return keepalive.EnforcementPolicy{
		MinTime:             serverCfg.KeepaliveMinTime,
		PermitWithoutStream: true,
	}
}

func (s *Service) Health(context.Context, *agentcontracts.HealthRequest) (*agentcontracts.HealthResponse, error) {
	capabilities := make([]string, 0, 3)
	if s != nil && s.hostExecutor != nil {
		capabilities = append(capabilities, "runtime_host_execution")
	}
	if s != nil && s.sandboxExecutor != nil {
		capabilities = append(capabilities, "checker_runner")
	}
	if s != nil && s.hostExecutor != nil {
		capabilities = append(capabilities, "interactive_exec")
	}
	return &agentcontracts.HealthResponse{
		Ready:        s != nil && s.hostExecutor != nil && s.sandboxExecutor != nil,
		Capabilities: capabilities,
	}, nil
}

func (s *Service) CreateNetwork(ctx context.Context, req *agentcontracts.CreateNetworkRequest) (*agentcontracts.CreateNetworkResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	networkID, err := executor.CreateNetwork(ctx, req.Name, req.Labels, req.Internal, req.AllowExisting, req.Subnet)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.CreateNetworkResponse{NetworkID: networkID}, nil
}

func (s *Service) ListNetworkSubnets(ctx context.Context, _ *agentcontracts.ListNetworkSubnetsRequest) (*agentcontracts.ListNetworkSubnetsResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	subnets, err := executor.ListNetworkSubnets(ctx)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ListNetworkSubnetsResponse{Subnets: subnets}, nil
}

func (s *Service) CreateContainer(ctx context.Context, req *agentcontracts.CreateContainerRequest) (*agentcontracts.CreateContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	containerID, err := executor.CreateContainer(ctx, req.Config)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.CreateContainerResponse{ContainerID: containerID}, nil
}

func (s *Service) ResolveServicePort(ctx context.Context, req *agentcontracts.ResolveServicePortRequest) (*agentcontracts.ResolveServicePortResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	port, err := executor.ResolveServicePort(ctx, req.ImageRef, req.PreferredPort)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ResolveServicePortResponse{Port: port}, nil
}

func (s *Service) ConnectContainerToNetwork(ctx context.Context, req *agentcontracts.ConnectContainerToNetworkRequest) (*agentcontracts.ConnectContainerToNetworkResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.ConnectContainerToNetwork(ctx, req.ContainerID, req.NetworkName); err != nil {
		return nil, err
	}
	return &agentcontracts.ConnectContainerToNetworkResponse{}, nil
}

func (s *Service) InspectContainerNetworkIPs(ctx context.Context, req *agentcontracts.InspectContainerNetworkIPsRequest) (*agentcontracts.InspectContainerNetworkIPsResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	networkIPs, err := executor.InspectContainerNetworkIPs(ctx, req.ContainerID)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.InspectContainerNetworkIPsResponse{NetworkIPs: networkIPs}, nil
}

func (s *Service) StartContainer(ctx context.Context, req *agentcontracts.StartContainerRequest) (*agentcontracts.StartContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.StartContainer(ctx, req.ContainerID); err != nil {
		return nil, err
	}
	return &agentcontracts.StartContainerResponse{}, nil
}

func (s *Service) StopContainer(ctx context.Context, req *agentcontracts.StopContainerRequest) (*agentcontracts.StopContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.StopContainer(ctx, req.ContainerID, time.Duration(req.TimeoutNanos)); err != nil {
		return nil, err
	}
	return &agentcontracts.StopContainerResponse{}, nil
}

func (s *Service) RemoveContainer(ctx context.Context, req *agentcontracts.RemoveContainerRequest) (*agentcontracts.RemoveContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.RemoveContainer(ctx, req.ContainerID, req.Force); err != nil {
		return nil, err
	}
	return &agentcontracts.RemoveContainerResponse{}, nil
}

func (s *Service) RemoveNetwork(ctx context.Context, req *agentcontracts.RemoveNetworkRequest) (*agentcontracts.RemoveNetworkResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.RemoveNetwork(ctx, req.NetworkID); err != nil {
		return nil, err
	}
	return &agentcontracts.RemoveNetworkResponse{}, nil
}

func (s *Service) ApplyACLRules(ctx context.Context, req *agentcontracts.ApplyACLRulesRequest) (*agentcontracts.ApplyACLRulesResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.ApplyACLRules(ctx, req.Rules); err != nil {
		return nil, err
	}
	return &agentcontracts.ApplyACLRulesResponse{}, nil
}

func (s *Service) ApplyACL(ctx context.Context, req *agentcontracts.ApplyACLRequest) (*agentcontracts.ApplyACLResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.ApplyACL(ctx, req.Handle, req.Rules); err != nil {
		return nil, err
	}
	return &agentcontracts.ApplyACLResponse{}, nil
}

func (s *Service) RemoveACLRules(ctx context.Context, req *agentcontracts.RemoveACLRulesRequest) (*agentcontracts.RemoveACLRulesResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.RemoveACLRules(ctx, req.Rules); err != nil {
		return nil, err
	}
	return &agentcontracts.RemoveACLRulesResponse{}, nil
}

func (s *Service) RemoveACL(ctx context.Context, req *agentcontracts.RemoveACLRequest) (*agentcontracts.RemoveACLResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.RemoveACL(ctx, req.Handle); err != nil {
		return nil, err
	}
	return &agentcontracts.RemoveACLResponse{}, nil
}

func (s *Service) WriteFileToContainer(ctx context.Context, req *agentcontracts.WriteFileToContainerRequest) (*agentcontracts.WriteFileToContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.WriteFileToContainer(ctx, req.ContainerID, req.FilePath, req.Content); err != nil {
		return nil, err
	}
	return &agentcontracts.WriteFileToContainerResponse{}, nil
}

func (s *Service) ReadFileFromContainer(ctx context.Context, req *agentcontracts.ReadFileFromContainerRequest) (*agentcontracts.ReadFileFromContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	content, err := executor.ReadFileFromContainer(ctx, req.ContainerID, req.FilePath, req.Limit)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ReadFileFromContainerResponse{Content: content}, nil
}

func (s *Service) ListDirectoryFromContainer(ctx context.Context, req *agentcontracts.ListDirectoryFromContainerRequest) (*agentcontracts.ListDirectoryFromContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	entries, err := executor.ListDirectoryFromContainer(ctx, req.ContainerID, req.DirPath, req.Limit)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ListDirectoryFromContainerResponse{Entries: entries}, nil
}

func (s *Service) ExecContainerCommand(ctx context.Context, req *agentcontracts.ExecContainerCommandRequest) (*agentcontracts.ExecContainerCommandResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	output, err := executor.ExecContainerCommand(ctx, req.ContainerID, req.Command, req.Stdin, req.Limit)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ExecContainerCommandResponse{Output: output}, nil
}

func (s *Service) InspectImageSize(ctx context.Context, req *agentcontracts.InspectImageSizeRequest) (*agentcontracts.InspectImageSizeResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	size, err := executor.InspectImageSize(ctx, req.ImageRef)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.InspectImageSizeResponse{Size: size}, nil
}

func (s *Service) RemoveImage(ctx context.Context, req *agentcontracts.RemoveImageRequest) (*agentcontracts.RemoveImageResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	if err := executor.RemoveImage(ctx, req.ImageRef); err != nil {
		return nil, err
	}
	return &agentcontracts.RemoveImageResponse{}, nil
}

func (s *Service) ListManagedContainers(ctx context.Context, _ *agentcontracts.ListManagedContainersRequest) (*agentcontracts.ListManagedContainersResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	containers, err := executor.ListManagedContainers(ctx)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ListManagedContainersResponse{Containers: containers}, nil
}

func (s *Service) InspectManagedContainer(ctx context.Context, req *agentcontracts.InspectManagedContainerRequest) (*agentcontracts.InspectManagedContainerResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	state, err := executor.InspectManagedContainer(ctx, req.ContainerID)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.InspectManagedContainerResponse{State: state}, nil
}

func (s *Service) ListManagedContainerStats(ctx context.Context, _ *agentcontracts.ListManagedContainerStatsRequest) (*agentcontracts.ListManagedContainerStatsResponse, error) {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return nil, err
	}
	stats, err := executor.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.ListManagedContainerStatsResponse{Stats: stats}, nil
}

func (s *Service) RunSandboxExec(ctx context.Context, req *agentcontracts.RunSandboxExecRequest) (*agentcontracts.RunSandboxExecResponse, error) {
	executor, err := s.requireSandboxExecutor()
	if err != nil {
		return nil, err
	}
	result, err := executor.RunSandboxExec(ctx, req.Job)
	if err != nil {
		return nil, err
	}
	return &agentcontracts.RunSandboxExecResponse{Result: result}, nil
}

func (s *Service) ExecContainerInteractive(stream agentcontracts.RuntimeAgent_ExecContainerInteractiveServer) error {
	executor, err := s.requireHostExecutor()
	if err != nil {
		return err
	}

	openReq, err := stream.Recv()
	if err != nil {
		return err
	}
	if openReq.Open == nil {
		return errors.New("interactive exec open frame is required")
	}

	stdinReader, stdinWriter := io.Pipe()
	defer func() { _ = stdinWriter.Close() }()

	resultCh := make(chan error, 1)
	go func() {
		resultCh <- executor.ExecContainerInteractive(
			stream.Context(),
			openReq.Open.ContainerID,
			openReq.Open.Command,
			stdinReader,
			&interactiveStreamWriter{stream: stream},
		)
	}()

	for {
		req, recvErr := stream.Recv()
		if recvErr != nil {
			if errors.Is(recvErr, io.EOF) {
				_ = stdinWriter.Close()
				return <-resultCh
			}
			_ = stdinWriter.CloseWithError(recvErr)
			<-resultCh
			return recvErr
		}
		if len(req.Stdin) == 0 {
			continue
		}
		if _, err := stdinWriter.Write(req.Stdin); err != nil {
			<-resultCh
			return err
		}
	}
}

func (s *Service) requireHostExecutor() (runtimeports.RuntimeHostExecutor, error) {
	if s == nil || s.hostExecutor == nil {
		return nil, errRuntimeExecutorUnavailable
	}
	return s.hostExecutor, nil
}

func (s *Service) requireSandboxExecutor() (runtimeports.SandboxExecutor, error) {
	if s == nil || s.sandboxExecutor == nil {
		return nil, errSandboxExecutorUnavailable
	}
	return s.sandboxExecutor, nil
}

type interactiveStreamWriter struct {
	stream agentcontracts.RuntimeAgent_ExecContainerInteractiveServer
	mu     sync.Mutex
}

func (w *interactiveStreamWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	chunk := append([]byte(nil), p...)
	if err := w.stream.Send(&agentcontracts.ExecContainerInteractiveResponse{Stdout: chunk}); err != nil {
		return 0, err
	}
	return len(p), nil
}
