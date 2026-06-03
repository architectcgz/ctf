package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
)

func assertHasRoute(t *testing.T, router *gin.Engine, method, path string) {
	t.Helper()

	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			return
		}
	}

	t.Fatalf("route not found: %s %s", method, path)
}

func assertRouteHandlerContains(t *testing.T, router *gin.Engine, method, path, want string) {
	t.Helper()

	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			if !strings.Contains(route.Handler, want) {
				t.Fatalf("route handler for %s %s = %s, want substring %s", method, path, route.Handler, want)
			}
			return
		}
	}

	t.Fatalf("route not found: %s %s", method, path)
}

func assertRouteMissing(t *testing.T, router *gin.Engine, method, path string) {
	t.Helper()

	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			t.Fatalf("unexpected route registered: %s %s", method, path)
		}
	}
}

func assertFieldType(t *testing.T, structType reflect.Type, fieldName string, want reflect.Type) {
	t.Helper()

	field, ok := structType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s missing field %s", structType.Name(), fieldName)
	}
	if field.Type != want {
		t.Fatalf("%s.%s type = %s, want %s", structType.Name(), fieldName, field.Type, want)
	}
}

func assertNoField(t *testing.T, structType reflect.Type, fieldName string) {
	t.Helper()

	if _, ok := structType.FieldByName(fieldName); ok {
		t.Fatalf("%s unexpectedly exposes field %s", structType.Name(), fieldName)
	}
}

func assertFunctionParamType(t *testing.T, fnType reflect.Type, index int, want reflect.Type) {
	t.Helper()

	if fnType.Kind() != reflect.Func {
		t.Fatalf("expected function type, got %s", fnType.Kind())
	}
	if index < 0 || index >= fnType.NumIn() {
		t.Fatalf("function has %d params, index %d out of range", fnType.NumIn(), index)
	}
	if got := fnType.In(index); got != want {
		t.Fatalf("function param %d type = %s, want %s", index, got, want)
	}
}

func newAppTestDependencies(t *testing.T) (*config.Config, *gorm.DB, *redislib.Client) {
	t.Helper()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = cache.Close()
	})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}

	db := openInternalAppTestSQLite(t, "router.sqlite")

	return newPracticeFlowTestConfig(t), db, cache
}

func writeRouterRemoteAgentClientTLSFiles(t *testing.T) (string, string, string) {
	t.Helper()

	dir := t.TempDir()
	certPEM, keyPEM := newRouterSelfSignedClientCertificatePEM(t, "runtime-agent.local")
	caFile := filepath.Join(dir, "ca.pem")
	certFile := filepath.Join(dir, "client.pem")
	keyFile := filepath.Join(dir, "client-key.pem")

	for _, file := range []struct {
		path string
		data []byte
	}{
		{path: caFile, data: certPEM},
		{path: certFile, data: certPEM},
		{path: keyFile, data: keyPEM},
	} {
		if err := os.WriteFile(file.path, file.data, 0o600); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", file.path, err)
		}
	}

	return caFile, certFile, keyFile
}

func newRouterSelfSignedClientCertificatePEM(t *testing.T, commonName string) ([]byte, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IsCA:        true,
		DNSNames:    []string{commonName},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
}
