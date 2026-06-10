package agentclient_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	"ctf-platform/internal/module/container_runtime/agentcontracts"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	"ctf-platform/internal/module/container_runtime/infrastructure/agentclient"
	"ctf-platform/internal/module/container_runtime/infrastructure/agentserver"
	contestports "ctf-platform/internal/module/contest/ports"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestRemoteExecutionBridgeDelegatesRuntimeAndCheckerCallsOverMTLS(t *testing.T) {
	t.Parallel()

	serverTLS, clientTLS := newMutualTLSConfigs(t)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	t.Cleanup(func() { _ = lis.Close() })

	hostExecutor := &fakeRuntimeHostExecutor{
		createNetworkID: "network-123",
	}
	checkerRunner := &fakeCheckerRunner{
		result: contestports.CheckerRunResult{
			Status:   contestports.CheckerRunStatusOK,
			Reason:   contestports.CheckerReasonPassed,
			ExitCode: 0,
			Stdout:   "checker-ok",
			Duration: 250 * time.Millisecond,
		},
	}

	server := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(serverTLS)),
		grpc.ForceServerCodec(agentcontracts.JSONCodec()),
	)
	agentcontracts.RegisterRuntimeAgentService(server, agentserver.NewService(hostExecutor, checkerRunner))
	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(server.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		lis.Addr().String(),
		grpc.WithTransportCredentials(credentials.NewTLS(clientTLS)),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(agentcontracts.JSONCodec())),
	)
	if err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	contractsClient := agentcontracts.NewRuntimeAgentClient(conn)
	healthResp, err := contractsClient.Health(ctx, &agentcontracts.HealthRequest{})
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if !healthResp.Ready {
		t.Fatalf("expected health ready, got %+v", healthResp)
	}

	bridge := agentclient.New(conn)

	networkID, err := bridge.CreateNetwork(ctx, "ctf-net", map[string]string{"scope": "test"}, true, true, "10.10.0.0/24")
	if err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}
	if networkID != "network-123" {
		t.Fatalf("CreateNetwork() = %q, want %q", networkID, "network-123")
	}
	if hostExecutor.lastCreateNetworkName != "ctf-net" {
		t.Fatalf("expected remote host executor to receive network name, got %q", hostExecutor.lastCreateNetworkName)
	}

	aclHandle := &runtimecontracts.InstanceRuntimeACLHandle{Chain: "CTF-INS-77"}
	aclRules := []runtimecontracts.InstanceRuntimeACLRule{
		{SourceIP: "10.10.0.2", TargetIP: "10.10.0.3", Action: "allow", Protocol: "tcp", Ports: []int{8080}},
	}
	if err := bridge.ApplyACL(ctx, aclHandle, aclRules); err != nil {
		t.Fatalf("ApplyACL() error = %v", err)
	}
	if hostExecutor.appliedACLHandle == nil || hostExecutor.appliedACLHandle.Chain != "CTF-INS-77" {
		t.Fatalf("expected remote host executor to receive acl handle, got %+v", hostExecutor.appliedACLHandle)
	}
	if len(hostExecutor.appliedACLRules) != 1 || hostExecutor.appliedACLRules[0].TargetIP != "10.10.0.3" {
		t.Fatalf("expected remote host executor to receive acl rules, got %+v", hostExecutor.appliedACLRules)
	}
	if err := bridge.RemoveACL(ctx, aclHandle); err != nil {
		t.Fatalf("RemoveACL() error = %v", err)
	}
	if hostExecutor.removedACLHandle == nil || hostExecutor.removedACLHandle.Chain != "CTF-INS-77" {
		t.Fatalf("expected remote host executor to receive acl removal handle, got %+v", hostExecutor.removedACLHandle)
	}

	result, err := bridge.RunChecker(ctx, contestports.CheckerRunJob{
		Runtime: "python3",
		Image:   "python:3.12-alpine",
		Entry:   "/checker/main.py",
		Env: map[string]string{
			"FLAG": "ctf{remote}",
		},
		Files: []contestports.CheckerRunFile{
			{Path: "main.py", Content: []byte("print('ok')"), Mode: 0o755},
		},
		Timeout: 3 * time.Second,
		Metadata: contestports.CheckerRunMetadata{
			ContestID:   7,
			ServiceID:   9,
			TeamID:      11,
			RoundNumber: 3,
		},
	})
	if err != nil {
		t.Fatalf("RunChecker() error = %v", err)
	}
	if result.Status != contestports.CheckerRunStatusOK || result.Stdout != "checker-ok" {
		t.Fatalf("unexpected checker result: %+v", result)
	}
	if checkerRunner.lastJob.Metadata.ContestID != 7 || checkerRunner.lastJob.Metadata.TeamID != 11 {
		t.Fatalf("expected remote checker runner to receive job metadata, got %+v", checkerRunner.lastJob.Metadata)
	}
}

type fakeRuntimeHostExecutor struct {
	createNetworkID       string
	lastCreateNetworkName string
	appliedACLHandle      *runtimecontracts.InstanceRuntimeACLHandle
	removedACLHandle      *runtimecontracts.InstanceRuntimeACLHandle
	appliedACLRules       []runtimecontracts.InstanceRuntimeACLRule
}

func (f *fakeRuntimeHostExecutor) CreateNetwork(_ context.Context, name string, _ map[string]string, _ bool, _ bool, _ string) (string, error) {
	f.lastCreateNetworkName = name
	return f.createNetworkID, nil
}

func (*fakeRuntimeHostExecutor) ListNetworkSubnets(context.Context) ([]string, error) {
	return []string{"10.10.0.0/24"}, nil
}

func (*fakeRuntimeHostExecutor) CreateContainer(context.Context, *runtimecontracts.ContainerConfig) (string, error) {
	return "container-123", nil
}

func (*fakeRuntimeHostExecutor) ResolveServicePort(context.Context, string, int) (int, error) {
	return 8080, nil
}

func (*fakeRuntimeHostExecutor) ConnectContainerToNetwork(context.Context, string, string) error {
	return nil
}

func (*fakeRuntimeHostExecutor) InspectContainerNetworkIPs(context.Context, string) (map[string]string, error) {
	return map[string]string{"ctf-net": "10.10.0.2"}, nil
}

func (*fakeRuntimeHostExecutor) StartContainer(context.Context, string) error {
	return nil
}

func (*fakeRuntimeHostExecutor) StopContainer(context.Context, string, time.Duration) error {
	return nil
}

func (*fakeRuntimeHostExecutor) RemoveContainer(context.Context, string, bool) error {
	return nil
}

func (*fakeRuntimeHostExecutor) RemoveNetwork(context.Context, string) error {
	return nil
}

func (f *fakeRuntimeHostExecutor) ApplyACLRules(_ context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeHostExecutor) ApplyACL(_ context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	f.appliedACLHandle = handle
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (*fakeRuntimeHostExecutor) RemoveACLRules(context.Context, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (f *fakeRuntimeHostExecutor) RemoveACL(_ context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle) error {
	f.removedACLHandle = handle
	return nil
}

func (*fakeRuntimeHostExecutor) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (*fakeRuntimeHostExecutor) ReadFileFromContainer(context.Context, string, string, int64) ([]byte, error) {
	return []byte("hello"), nil
}

func (*fakeRuntimeHostExecutor) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimecontracts.ContainerDirectoryEntry, error) {
	return []runtimecontracts.ContainerDirectoryEntry{{Name: "main.py", Type: "file", Size: 64}}, nil
}

func (*fakeRuntimeHostExecutor) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return []byte("command-ok"), nil
}

func (*fakeRuntimeHostExecutor) InspectImageSize(context.Context, string) (int64, error) {
	return 1024, nil
}

func (*fakeRuntimeHostExecutor) RemoveImage(context.Context, string) error {
	return nil
}

func (*fakeRuntimeHostExecutor) ListManagedContainers(context.Context) ([]runtimecontracts.ManagedContainer, error) {
	return []runtimecontracts.ManagedContainer{{ID: "container-123", Name: "ctf-container", CreatedAt: time.Unix(0, 0).UTC()}}, nil
}

func (*fakeRuntimeHostExecutor) InspectManagedContainer(context.Context, string) (*runtimecontracts.ManagedContainerState, error) {
	return &runtimecontracts.ManagedContainerState{ID: "container-123", Exists: true, Running: true, Status: "running"}, nil
}

func (*fakeRuntimeHostExecutor) ListManagedContainerStats(context.Context) ([]runtimecontracts.ManagedContainerStat, error) {
	return []runtimecontracts.ManagedContainerStat{{ContainerID: "container-123", ContainerName: "ctf-container", CPUPercent: 12.5}}, nil
}

func (*fakeRuntimeHostExecutor) ExecContainerInteractive(_ context.Context, _ string, _ []string, stdin io.Reader, stdout io.Writer) error {
	if stdin == nil || stdout == nil {
		return nil
	}
	payload, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	_, err = stdout.Write(payload)
	return err
}

type fakeCheckerRunner struct {
	lastJob contestports.CheckerRunJob
	result  contestports.CheckerRunResult
}

func (f *fakeCheckerRunner) RunChecker(_ context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	f.lastJob = job
	return f.result, nil
}

func newMutualTLSConfigs(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()

	caCertPEM, caKeyPEM, caCert := newCertificateAuthority(t)
	serverCert := newSignedLeafCertificate(t, caCert, caKeyPEM, true, "runtime-agent.local", "127.0.0.1")
	clientCert := newSignedLeafCertificate(t, caCert, caKeyPEM, false, "runtime-api.local")

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPEM) {
		t.Fatal("AppendCertsFromPEM(ca) = false")
	}

	serverPair, err := tls.X509KeyPair(serverCert.certPEM, serverCert.keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair(server) error = %v", err)
	}
	clientPair, err := tls.X509KeyPair(clientCert.certPEM, clientCert.keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair(client) error = %v", err)
	}

	return &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{serverPair},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}, &tls.Config{
			MinVersion:   tls.VersionTLS13,
			ServerName:   "runtime-agent.local",
			Certificates: []tls.Certificate{clientPair},
			RootCAs:      certPool,
		}
}

type pemCertificate struct {
	certPEM []byte
	keyPEM  []byte
}

func newCertificateAuthority(t *testing.T) ([]byte, []byte, *x509.Certificate) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey(ca) error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "ctf-runtime-agent-ca",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate(ca) error = %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		t.Fatalf("ParseCertificate(ca) error = %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}),
		cert
}

func newSignedLeafCertificate(t *testing.T, caCert *x509.Certificate, caKeyPEM []byte, server bool, names ...string) pemCertificate {
	t.Helper()

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		t.Fatal("failed to decode CA private key")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		t.Fatalf("ParsePKCS1PrivateKey(ca) error = %v", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey(leaf) error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: names[0],
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	if server {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	for _, name := range names {
		if ip := net.ParseIP(name); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
			continue
		}
		template.DNSNames = append(template.DNSNames, name)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("CreateCertificate(leaf) error = %v", err)
	}

	return pemCertificate{
		certPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}),
		keyPEM:  pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}),
	}
}
