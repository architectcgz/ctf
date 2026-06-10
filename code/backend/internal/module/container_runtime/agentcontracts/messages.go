package agentcontracts

import (
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

type HealthRequest struct{}

type HealthResponse struct {
	Ready        bool     `json:"ready"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type CreateNetworkRequest struct {
	Name          string            `json:"name"`
	Labels        map[string]string `json:"labels,omitempty"`
	Internal      bool              `json:"internal"`
	AllowExisting bool              `json:"allow_existing"`
	Subnet        string            `json:"subnet,omitempty"`
}

type CreateNetworkResponse struct {
	NetworkID string `json:"network_id"`
}

type ListNetworkSubnetsRequest struct{}

type ListNetworkSubnetsResponse struct {
	Subnets []string `json:"subnets"`
}

type CreateContainerRequest struct {
	Config *runtimecontracts.ContainerConfig `json:"config,omitempty"`
}

type CreateContainerResponse struct {
	ContainerID string `json:"container_id"`
}

type ResolveServicePortRequest struct {
	ImageRef      string `json:"image_ref"`
	PreferredPort int    `json:"preferred_port"`
}

type ResolveServicePortResponse struct {
	Port int `json:"port"`
}

type ConnectContainerToNetworkRequest struct {
	ContainerID string `json:"container_id"`
	NetworkName string `json:"network_name"`
}

type ConnectContainerToNetworkResponse struct{}

type InspectContainerNetworkIPsRequest struct {
	ContainerID string `json:"container_id"`
}

type InspectContainerNetworkIPsResponse struct {
	NetworkIPs map[string]string `json:"network_ips,omitempty"`
}

type StartContainerRequest struct {
	ContainerID string `json:"container_id"`
}

type StartContainerResponse struct{}

type StopContainerRequest struct {
	ContainerID  string `json:"container_id"`
	TimeoutNanos int64  `json:"timeout_nanos"`
}

type StopContainerResponse struct{}

type RemoveContainerRequest struct {
	ContainerID string `json:"container_id"`
	Force       bool   `json:"force"`
}

type RemoveContainerResponse struct{}

type RemoveNetworkRequest struct {
	NetworkID string `json:"network_id"`
}

type RemoveNetworkResponse struct{}

type ApplyACLRulesRequest struct {
	Rules []runtimecontracts.InstanceRuntimeACLRule `json:"rules,omitempty"`
}

type ApplyACLRulesResponse struct{}

type ApplyACLRequest struct {
	Handle *runtimecontracts.InstanceRuntimeACLHandle `json:"handle,omitempty"`
	Rules  []runtimecontracts.InstanceRuntimeACLRule  `json:"rules,omitempty"`
}

type ApplyACLResponse struct{}

type RemoveACLRulesRequest struct {
	Rules []runtimecontracts.InstanceRuntimeACLRule `json:"rules,omitempty"`
}

type RemoveACLRulesResponse struct{}

type RemoveACLRequest struct {
	Handle *runtimecontracts.InstanceRuntimeACLHandle `json:"handle,omitempty"`
}

type RemoveACLResponse struct{}

type WriteFileToContainerRequest struct {
	ContainerID string `json:"container_id"`
	FilePath    string `json:"file_path"`
	Content     []byte `json:"content,omitempty"`
}

type WriteFileToContainerResponse struct{}

type ReadFileFromContainerRequest struct {
	ContainerID string `json:"container_id"`
	FilePath    string `json:"file_path"`
	Limit       int64  `json:"limit"`
}

type ReadFileFromContainerResponse struct {
	Content []byte `json:"content,omitempty"`
}

type ListDirectoryFromContainerRequest struct {
	ContainerID string `json:"container_id"`
	DirPath     string `json:"dir_path"`
	Limit       int    `json:"limit"`
}

type ListDirectoryFromContainerResponse struct {
	Entries []runtimecontracts.ContainerDirectoryEntry `json:"entries,omitempty"`
}

type ExecContainerCommandRequest struct {
	ContainerID string   `json:"container_id"`
	Command     []string `json:"command,omitempty"`
	Stdin       []byte   `json:"stdin,omitempty"`
	Limit       int64    `json:"limit"`
}

type ExecContainerCommandResponse struct {
	Output []byte `json:"output,omitempty"`
}

type InspectImageSizeRequest struct {
	ImageRef string `json:"image_ref"`
}

type InspectImageSizeResponse struct {
	Size int64 `json:"size"`
}

type RemoveImageRequest struct {
	ImageRef string `json:"image_ref"`
}

type RemoveImageResponse struct{}

type ListManagedContainersRequest struct{}

type ListManagedContainersResponse struct {
	Containers []runtimecontracts.ManagedContainer `json:"containers,omitempty"`
}

type InspectManagedContainerRequest struct {
	ContainerID string `json:"container_id"`
}

type InspectManagedContainerResponse struct {
	State *runtimecontracts.ManagedContainerState `json:"state,omitempty"`
}

type ListManagedContainerStatsRequest struct{}

type ListManagedContainerStatsResponse struct {
	Stats []runtimecontracts.ManagedContainerStat `json:"stats,omitempty"`
}

type RunSandboxExecRequest struct {
	Job runtimeports.SandboxExecJob `json:"job"`
}

type RunSandboxExecResponse struct {
	Result runtimeports.SandboxExecResult `json:"result"`
}

type ExecContainerInteractiveOpen struct {
	ContainerID string   `json:"container_id"`
	Command     []string `json:"command,omitempty"`
}

type ExecContainerInteractiveRequest struct {
	Open  *ExecContainerInteractiveOpen `json:"open,omitempty"`
	Stdin []byte                        `json:"stdin,omitempty"`
}

type ExecContainerInteractiveResponse struct {
	Stdout []byte `json:"stdout,omitempty"`
}
