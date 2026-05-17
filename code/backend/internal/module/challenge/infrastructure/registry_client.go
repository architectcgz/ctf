package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"ctf-platform/internal/module/challenge/domain"
)

var registryManifestAcceptHeader = strings.Join([]string{
	"application/vnd.docker.distribution.manifest.v2+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
	"application/vnd.oci.image.manifest.v1+json",
	"application/vnd.oci.image.index.v1+json",
}, ", ")

type RegistryClientConfig struct {
	Scheme        string
	Server        string
	AccessServer  string
	Username      string
	Password      string
	IdentityToken string
}

type RegistryClient struct {
	config RegistryClientConfig
	client *http.Client
}

func NewRegistryClient(config RegistryClientConfig, client *http.Client) *RegistryClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &RegistryClient{config: config, client: client}
}

func (c *RegistryClient) CheckManifest(ctx context.Context, imageRef string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("registry client is not configured")
	}
	scheme := strings.TrimSpace(c.config.Scheme)
	if scheme == "" {
		scheme = "https"
	}
	server := strings.Trim(strings.TrimSpace(c.config.Server), "/")
	if server == "" {
		return "", fmt.Errorf("registry server is required")
	}
	accessServer := normalizeRegistryServerEndpoint(c.config.AccessServer)
	if accessServer == "" {
		accessServer = normalizeRegistryServerEndpoint(server)
	}

	name, tag, err := domain.SplitImageRef(imageRef)
	if err != nil {
		return "", err
	}
	repository := strings.TrimPrefix(name, server+"/")
	if repository == name {
		return "", fmt.Errorf("image ref %q does not belong to registry %q", imageRef, server)
	}
	manifestURL := url.URL{
		Scheme: scheme,
		Host:   accessServer,
		Path:   fmt.Sprintf("/v2/%s/manifests/%s", repository, tag),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, manifestURL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", registryManifestAcceptHeader)
	if strings.TrimSpace(c.config.IdentityToken) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.config.IdentityToken))
	} else if strings.TrimSpace(c.config.Username) != "" || strings.TrimSpace(c.config.Password) != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("check registry manifest %s: %w", imageRef, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("check registry manifest %s: status %d", imageRef, resp.StatusCode)
	}
	digest := strings.TrimSpace(resp.Header.Get("Docker-Content-Digest"))
	if digest == "" {
		return "", fmt.Errorf("check registry manifest %s: missing Docker-Content-Digest", imageRef)
	}
	return digest, nil
}

func normalizeRegistryServerEndpoint(server string) string {
	normalized := strings.TrimSpace(server)
	normalized = strings.TrimPrefix(normalized, "https://")
	normalized = strings.TrimPrefix(normalized, "http://")
	normalized = strings.Trim(normalized, "/")
	if slash := strings.Index(normalized, "/"); slash >= 0 {
		normalized = normalized[:slash]
	}
	return normalized
}
