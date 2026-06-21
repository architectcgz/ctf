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
	"ctf-platform/internal/module/container_runtime/agentcontracts"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type Client struct {
	conn   *grpc.ClientConn
	client agentcontracts.RuntimeAgentClient
}

var _ runtimeports.RuntimeHostExecutor = (*Client)(nil)
var _ runtimeports.SandboxExecutor = (*Client)(nil)

func runtimeAgentClientKeepaliveParameters(cfg config.RuntimeAgentConfig) keepalive.ClientParameters {
	return keepalive.ClientParameters{
		Time:                cfg.KeepaliveTime,
		Timeout:             cfg.KeepaliveTimeout,
		PermitWithoutStream: true,
	}
}

func New(conn *grpc.ClientConn) *Client {
	if conn == nil {
		return &Client{}
	}
	return &Client{
		conn:   conn,
		client: agentcontracts.NewRuntimeAgentClient(conn),
	}
}

func DialContext(ctx context.Context, cfg config.RuntimeAgentConfig) (*Client, error) {
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
		grpc.WithKeepaliveParams(runtimeAgentClientKeepaliveParameters(cfg)),
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

func (c *Client) Close(context.Context) error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Health(ctx context.Context) (*agentcontracts.HealthResponse, error) {
	return c.requireClient().Health(ctx, &agentcontracts.HealthRequest{})
}

func (c *Client) CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool, subnet string) (string, error) {
	resp, err := c.requireClient().CreateNetwork(ctx, &agentcontracts.CreateNetworkRequest{
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

func (c *Client) ListNetworkSubnets(ctx context.Context) ([]string, error) {
	resp, err := c.requireClient().ListNetworkSubnets(ctx, &agentcontracts.ListNetworkSubnetsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Subnets, nil
}

func (c *Client) CreateContainer(ctx context.Context, cfg *runtimecontracts.ContainerConfig) (string, error) {
	resp, err := c.requireClient().CreateContainer(ctx, &agentcontracts.CreateContainerRequest{Config: cfg})
	if err != nil {
		return "", err
	}
	return resp.ContainerID, nil
}

func (c *Client) ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error) {
	resp, err := c.requireClient().ResolveServicePort(ctx, &agentcontracts.ResolveServicePortRequest{ImageRef: imageRef, PreferredPort: preferredPort})
	if err != nil {
		return 0, err
	}
	return resp.Port, nil
}

func (c *Client) ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error {
	_, err := c.requireClient().ConnectContainerToNetwork(ctx, &agentcontracts.ConnectContainerToNetworkRequest{ContainerID: containerID, NetworkName: networkName})
	return err
}

func (c *Client) InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error) {
	resp, err := c.requireClient().InspectContainerNetworkIPs(ctx, &agentcontracts.InspectContainerNetworkIPsRequest{ContainerID: containerID})
	if err != nil {
		return nil, err
	}
	return resp.NetworkIPs, nil
}

func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	_, err := c.requireClient().StartContainer(ctx, &agentcontracts.StartContainerRequest{ContainerID: containerID})
	return err
}

func (c *Client) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	_, err := c.requireClient().StopContainer(ctx, &agentcontracts.StopContainerRequest{ContainerID: containerID, TimeoutNanos: timeout.Nanoseconds()})
	return err
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	_, err := c.requireClient().RemoveContainer(ctx, &agentcontracts.RemoveContainerRequest{ContainerID: containerID, Force: force})
	return err
}

func (c *Client) RemoveNetwork(ctx context.Context, networkID string) error {
	_, err := c.requireClient().RemoveNetwork(ctx, &agentcontracts.RemoveNetworkRequest{NetworkID: networkID})
	return err
}

func (c *Client) ApplyACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	_, err := c.requireClient().ApplyACLRules(ctx, &agentcontracts.ApplyACLRulesRequest{Rules: rules})
	return err
}

func (c *Client) ApplyACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	_, err := c.requireClient().ApplyACL(ctx, &agentcontracts.ApplyACLRequest{Handle: handle, Rules: rules})
	return err
}

func (c *Client) RemoveACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	_, err := c.requireClient().RemoveACLRules(ctx, &agentcontracts.RemoveACLRulesRequest{Rules: rules})
	return err
}

func (c *Client) RemoveACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle) error {
	_, err := c.requireClient().RemoveACL(ctx, &agentcontracts.RemoveACLRequest{Handle: handle})
	return err
}

func (c *Client) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	_, err := c.requireClient().WriteFileToContainer(ctx, &agentcontracts.WriteFileToContainerRequest{
		ContainerID: containerID,
		FilePath:    filePath,
		Content:     content,
	})
	return err
}

func (c *Client) ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error) {
	resp, err := c.requireClient().ReadFileFromContainer(ctx, &agentcontracts.ReadFileFromContainerRequest{
		ContainerID: containerID,
		FilePath:    filePath,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Content, nil
}

func (c *Client) ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimecontracts.ContainerDirectoryEntry, error) {
	resp, err := c.requireClient().ListDirectoryFromContainer(ctx, &agentcontracts.ListDirectoryFromContainerRequest{
		ContainerID: containerID,
		DirPath:     dirPath,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Entries, nil
}

func (c *Client) ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error) {
	resp, err := c.requireClient().ExecContainerCommand(ctx, &agentcontracts.ExecContainerCommandRequest{
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

func (c *Client) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	resp, err := c.requireClient().InspectImageSize(ctx, &agentcontracts.InspectImageSizeRequest{ImageRef: imageRef})
	if err != nil {
		return 0, err
	}
	return resp.Size, nil
}

func (c *Client) RemoveImage(ctx context.Context, imageRef string) error {
	_, err := c.requireClient().RemoveImage(ctx, &agentcontracts.RemoveImageRequest{ImageRef: imageRef})
	return err
}

func (c *Client) ListManagedContainers(ctx context.Context) ([]runtimecontracts.ManagedContainer, error) {
	resp, err := c.requireClient().ListManagedContainers(ctx, &agentcontracts.ListManagedContainersRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Containers, nil
}

func (c *Client) InspectManagedContainer(ctx context.Context, containerID string) (*runtimecontracts.ManagedContainerState, error) {
	resp, err := c.requireClient().InspectManagedContainer(ctx, &agentcontracts.InspectManagedContainerRequest{ContainerID: containerID})
	if err != nil {
		return nil, err
	}
	return resp.State, nil
}

func (c *Client) ListManagedContainerStats(ctx context.Context) ([]runtimecontracts.ManagedContainerStat, error) {
	resp, err := c.requireClient().ListManagedContainerStats(ctx, &agentcontracts.ListManagedContainerStatsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Stats, nil
}

func (c *Client) RunSandboxExec(ctx context.Context, job runtimeports.SandboxExecJob) (runtimeports.SandboxExecResult, error) {
	resp, err := c.requireClient().RunSandboxExec(ctx, &agentcontracts.RunSandboxExecRequest{Job: job})
	if err != nil {
		return runtimeports.SandboxExecResult{}, err
	}
	return resp.Result, nil
}

func (c *Client) ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error {
	stream, err := c.requireClient().ExecContainerInteractive(ctx)
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

func (c *Client) requireClient() agentcontracts.RuntimeAgentClient {
	if c != nil && c.client != nil {
		return c.client
	}
	return agentcontracts.NewRuntimeAgentClient(nil)
}
