package agentclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"time"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/module/runtime/agentcontracts"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Bridge struct {
	conn   *grpc.ClientConn
	client agentcontracts.RuntimeAgentClient
}

var _ runtimeports.RuntimeHostExecutor = (*Bridge)(nil)
var _ contestports.CheckerRunner = (*Bridge)(nil)

func New(conn *grpc.ClientConn) *Bridge {
	if conn == nil {
		return &Bridge{}
	}
	return &Bridge{
		conn:   conn,
		client: agentcontracts.NewRuntimeAgentClient(conn),
	}
}

func DialContext(ctx context.Context, cfg config.RuntimeAgentConfig) (*Bridge, error) {
	tlsConfig, err := LoadClientTLSConfig(cfg.CAFile, cfg.CertFile, cfg.KeyFile, cfg.ServerName)
	if err != nil {
		return nil, err
	}
	dialCtx := ctx
	cancel := func() {}
	if cfg.DialTimeout > 0 {
		dialCtx, cancel = context.WithTimeout(ctx, cfg.DialTimeout)
	}
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		cfg.Endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(agentcontracts.JSONCodec())),
	)
	if err != nil {
		return nil, err
	}
	return New(conn), nil
}

func LoadClientTLSConfig(caFile, certFile, keyFile, serverName string) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load runtime agent client key pair: %w", err)
	}
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read runtime agent ca: %w", err)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("append runtime agent ca pem failed")
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		ServerName:   serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      rootCAs,
	}, nil
}

func (b *Bridge) Close(context.Context) error {
	if b == nil || b.conn == nil {
		return nil
	}
	return b.conn.Close()
}

func (b *Bridge) Health(ctx context.Context) (*agentcontracts.HealthResponse, error) {
	return b.requireClient().Health(ctx, &agentcontracts.HealthRequest{})
}

func (b *Bridge) CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool, subnet string) (string, error) {
	resp, err := b.requireClient().CreateNetwork(ctx, &agentcontracts.CreateNetworkRequest{
		Name:          name,
		Labels:        labels,
		Internal:      internal,
		AllowExisting: allowExisting,
		Subnet:        subnet,
	})
	if err != nil {
		return "", err
	}
	return resp.NetworkID, nil
}

func (b *Bridge) ListNetworkSubnets(ctx context.Context) ([]string, error) {
	resp, err := b.requireClient().ListNetworkSubnets(ctx, &agentcontracts.ListNetworkSubnetsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Subnets, nil
}

func (b *Bridge) CreateContainer(ctx context.Context, cfg *runtimecontracts.ContainerConfig) (string, error) {
	resp, err := b.requireClient().CreateContainer(ctx, &agentcontracts.CreateContainerRequest{Config: cfg})
	if err != nil {
		return "", err
	}
	return resp.ContainerID, nil
}

func (b *Bridge) ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error) {
	resp, err := b.requireClient().ResolveServicePort(ctx, &agentcontracts.ResolveServicePortRequest{ImageRef: imageRef, PreferredPort: preferredPort})
	if err != nil {
		return 0, err
	}
	return resp.Port, nil
}

func (b *Bridge) ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error {
	_, err := b.requireClient().ConnectContainerToNetwork(ctx, &agentcontracts.ConnectContainerToNetworkRequest{ContainerID: containerID, NetworkName: networkName})
	return err
}

func (b *Bridge) InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error) {
	resp, err := b.requireClient().InspectContainerNetworkIPs(ctx, &agentcontracts.InspectContainerNetworkIPsRequest{ContainerID: containerID})
	if err != nil {
		return nil, err
	}
	return resp.NetworkIPs, nil
}

func (b *Bridge) StartContainer(ctx context.Context, containerID string) error {
	_, err := b.requireClient().StartContainer(ctx, &agentcontracts.StartContainerRequest{ContainerID: containerID})
	return err
}

func (b *Bridge) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	_, err := b.requireClient().StopContainer(ctx, &agentcontracts.StopContainerRequest{ContainerID: containerID, TimeoutNanos: timeout.Nanoseconds()})
	return err
}

func (b *Bridge) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	_, err := b.requireClient().RemoveContainer(ctx, &agentcontracts.RemoveContainerRequest{ContainerID: containerID, Force: force})
	return err
}

func (b *Bridge) RemoveNetwork(ctx context.Context, networkID string) error {
	_, err := b.requireClient().RemoveNetwork(ctx, &agentcontracts.RemoveNetworkRequest{NetworkID: networkID})
	return err
}

func (b *Bridge) ApplyACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	_, err := b.requireClient().ApplyACLRules(ctx, &agentcontracts.ApplyACLRulesRequest{Rules: rules})
	return err
}

func (b *Bridge) ApplyACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	_, err := b.requireClient().ApplyACL(ctx, &agentcontracts.ApplyACLRequest{Handle: handle, Rules: rules})
	return err
}

func (b *Bridge) RemoveACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	_, err := b.requireClient().RemoveACLRules(ctx, &agentcontracts.RemoveACLRulesRequest{Rules: rules})
	return err
}

func (b *Bridge) RemoveACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle) error {
	_, err := b.requireClient().RemoveACL(ctx, &agentcontracts.RemoveACLRequest{Handle: handle})
	return err
}

func (b *Bridge) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	_, err := b.requireClient().WriteFileToContainer(ctx, &agentcontracts.WriteFileToContainerRequest{
		ContainerID: containerID,
		FilePath:    filePath,
		Content:     content,
	})
	return err
}

func (b *Bridge) ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error) {
	resp, err := b.requireClient().ReadFileFromContainer(ctx, &agentcontracts.ReadFileFromContainerRequest{
		ContainerID: containerID,
		FilePath:    filePath,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Content, nil
}

func (b *Bridge) ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error) {
	resp, err := b.requireClient().ListDirectoryFromContainer(ctx, &agentcontracts.ListDirectoryFromContainerRequest{
		ContainerID: containerID,
		DirPath:     dirPath,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Entries, nil
}

func (b *Bridge) ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error) {
	resp, err := b.requireClient().ExecContainerCommand(ctx, &agentcontracts.ExecContainerCommandRequest{
		ContainerID: containerID,
		Command:     command,
		Stdin:       stdin,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Output, nil
}

func (b *Bridge) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	resp, err := b.requireClient().InspectImageSize(ctx, &agentcontracts.InspectImageSizeRequest{ImageRef: imageRef})
	if err != nil {
		return 0, err
	}
	return resp.Size, nil
}

func (b *Bridge) RemoveImage(ctx context.Context, imageRef string) error {
	_, err := b.requireClient().RemoveImage(ctx, &agentcontracts.RemoveImageRequest{ImageRef: imageRef})
	return err
}

func (b *Bridge) ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error) {
	resp, err := b.requireClient().ListManagedContainers(ctx, &agentcontracts.ListManagedContainersRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Containers, nil
}

func (b *Bridge) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	resp, err := b.requireClient().InspectManagedContainer(ctx, &agentcontracts.InspectManagedContainerRequest{ContainerID: containerID})
	if err != nil {
		return nil, err
	}
	return resp.State, nil
}

func (b *Bridge) ListManagedContainerStats(ctx context.Context) ([]runtimeports.ManagedContainerStat, error) {
	resp, err := b.requireClient().ListManagedContainerStats(ctx, &agentcontracts.ListManagedContainerStatsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Stats, nil
}

func (b *Bridge) RunChecker(ctx context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	resp, err := b.requireClient().RunChecker(ctx, &agentcontracts.RunCheckerRequest{Job: job})
	if err != nil {
		return contestports.CheckerRunResult{}, err
	}
	return resp.Result, nil
}

func (b *Bridge) ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error {
	stream, err := b.requireClient().ExecContainerInteractive(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(&agentcontracts.ExecContainerInteractiveRequest{
		Open: &agentcontracts.ExecContainerInteractiveOpen{
			ContainerID: containerID,
			Command:     command,
		},
	}); err != nil {
		return err
	}

	sendErrCh := make(chan error, 1)
	go func() {
		if stdin == nil {
			sendErrCh <- stream.CloseSend()
			return
		}
		buf := make([]byte, 32*1024)
		for {
			n, readErr := stdin.Read(buf)
			if n > 0 {
				payload := append([]byte(nil), buf[:n]...)
				if err := stream.Send(&agentcontracts.ExecContainerInteractiveRequest{Stdin: payload}); err != nil {
					sendErrCh <- err
					return
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					sendErrCh <- stream.CloseSend()
					return
				}
				sendErrCh <- readErr
				return
			}
		}
	}()

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return <-sendErrCh
			}
			return err
		}
		if stdout != nil && len(resp.Stdout) > 0 {
			if _, err := stdout.Write(resp.Stdout); err != nil {
				return err
			}
		}
	}
}

func (b *Bridge) requireClient() agentcontracts.RuntimeAgentClient {
	if b != nil && b.client != nil {
		return b.client
	}
	return agentcontracts.NewRuntimeAgentClient(nil)
}
