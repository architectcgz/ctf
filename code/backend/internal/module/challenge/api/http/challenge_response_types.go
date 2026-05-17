package http

import (
	"ctf-platform/internal/model"
	"time"
)

type PageResult[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"page_size"`
}

type TagResp struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ImageResp struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Tag           string     `json:"tag"`
	Description   string     `json:"description"`
	Size          int64      `json:"size"`
	SizeFormatted string     `json:"size_formatted"`
	Status        string     `json:"status"`
	Digest        string     `json:"digest,omitempty"`
	SourceType    string     `json:"source_type,omitempty"`
	BuildJobID    *int64     `json:"build_job_id,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type FlagResp struct {
	FlagType   string `json:"flag_type"`
	FlagRegex  string `json:"flag_regex,omitempty"`
	FlagPrefix string `json:"flag_prefix,omitempty"`
	Configured bool   `json:"configured"`
}

type ChallengeHintAdminResp struct {
	ID      int64  `json:"id"`
	Level   int    `json:"level"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
}

type ChallengeResp struct {
	ID              int64                     `json:"id"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	Category        string                    `json:"category"`
	Difficulty      string                    `json:"difficulty"`
	Points          int                       `json:"points"`
	ImageID         int64                     `json:"image_id"`
	AttachmentURL   string                    `json:"attachment_url,omitempty"`
	InstanceSharing model.InstanceSharing     `json:"instance_sharing"`
	Hints           []*ChallengeHintAdminResp `json:"hints,omitempty"`
	Status          string                    `json:"status"`
	CreatedBy       *int64                    `json:"created_by,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type ChallengeListItem struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"`
	Difficulty    string    `json:"difficulty"`
	Points        int       `json:"points"`
	SolvedCount   int64     `json:"solved_count"`
	TotalAttempts int64     `json:"total_attempts"`
	IsSolved      bool      `json:"is_solved"`
	CreatedAt     time.Time `json:"created_at"`
}

type ChallengeHintResp struct {
	ID      int64  `json:"id"`
	Level   int    `json:"level"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type ChallengeDetailResp struct {
	ID              int64                 `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Category        string                `json:"category"`
	Difficulty      string                `json:"difficulty"`
	Points          int                   `json:"points"`
	NeedTarget      bool                  `json:"need_target"`
	FlagType        string                `json:"flag_type"`
	InstanceSharing model.InstanceSharing `json:"instance_sharing"`
	AttachmentURL   string                `json:"attachment_url,omitempty"`
	Hints           []*ChallengeHintResp  `json:"hints"`
	SolvedCount     int64                 `json:"solved_count"`
	TotalAttempts   int64                 `json:"total_attempts"`
	IsSolved        bool                  `json:"is_solved"`
	CreatedAt       time.Time             `json:"created_at"`
}

type ChallengeImportCommitResp struct {
	Challenge *ChallengeResp `json:"challenge"`
}

type ChallengeSelfCheckStepResp struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type ChallengeSelfCheckPhaseResp struct {
	Passed    bool                         `json:"passed"`
	StartedAt time.Time                    `json:"started_at"`
	EndedAt   time.Time                    `json:"ended_at"`
	Steps     []ChallengeSelfCheckStepResp `json:"steps"`
}

type ChallengeSelfCheckRuntimeResp struct {
	Passed         bool                         `json:"passed"`
	StartedAt      time.Time                    `json:"started_at"`
	EndedAt        time.Time                    `json:"ended_at"`
	AccessURL      string                       `json:"access_url,omitempty"`
	ContainerCount int                          `json:"container_count"`
	NetworkCount   int                          `json:"network_count"`
	Steps          []ChallengeSelfCheckStepResp `json:"steps"`
}

type ChallengeSelfCheckResp struct {
	ChallengeID int64                         `json:"challenge_id"`
	Precheck    ChallengeSelfCheckPhaseResp   `json:"precheck"`
	Runtime     ChallengeSelfCheckRuntimeResp `json:"runtime"`
}

type ChallengePublishCheckJobResp struct {
	ID             int64                   `json:"id"`
	ChallengeID    int64                   `json:"challenge_id"`
	RequestedBy    int64                   `json:"requested_by"`
	Status         string                  `json:"status"`
	Active         bool                    `json:"active"`
	RequestSource  string                  `json:"request_source"`
	FailureSummary string                  `json:"failure_summary,omitempty"`
	StartedAt      *time.Time              `json:"started_at,omitempty"`
	FinishedAt     *time.Time              `json:"finished_at,omitempty"`
	PublishedAt    *time.Time              `json:"published_at,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	Result         *ChallengeSelfCheckResp `json:"result,omitempty"`
}

type ChallengeImportAttachmentResp struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ChallengeImportFlagResp struct {
	Type   string `json:"type"`
	Prefix string `json:"prefix,omitempty"`
}

type ChallengeImportRuntimeResp struct {
	Type     string `json:"type,omitempty"`
	ImageRef string `json:"image_ref,omitempty"`
}

type ChallengeImportTopologyExtensionResp struct {
	Source  string `json:"source,omitempty"`
	Enabled bool   `json:"enabled"`
}

type ChallengeImportTopologyNodeResp struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	ImageRef    string            `json:"image_ref,omitempty"`
	ServicePort int               `json:"service_port,omitempty"`
	InjectFlag  bool              `json:"inject_flag,omitempty"`
	Tier        string            `json:"tier,omitempty"`
	NetworkKeys []string          `json:"network_keys,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

type ChallengeImportTopologyResp struct {
	Source       string                            `json:"source,omitempty"`
	EntryNodeKey string                            `json:"entry_node_key"`
	Networks     []TopologyNetworkResp             `json:"networks,omitempty"`
	Nodes        []ChallengeImportTopologyNodeResp `json:"nodes"`
	Links        []TopologyLinkResp                `json:"links,omitempty"`
	Policies     []TopologyTrafficPolicyResp       `json:"policies,omitempty"`
}

type ChallengeImportExtensionsResp struct {
	Topology ChallengeImportTopologyExtensionResp `json:"topology"`
}

type ChallengeImportPreviewResp struct {
	ID            string                           `json:"id"`
	FileName      string                           `json:"file_name"`
	Slug          string                           `json:"slug"`
	Title         string                           `json:"title"`
	Description   string                           `json:"description"`
	Category      string                           `json:"category"`
	Difficulty    string                           `json:"difficulty"`
	Points        int                              `json:"points"`
	Attachments   []ChallengeImportAttachmentResp  `json:"attachments,omitempty"`
	Hints         []ChallengeHintAdminResp         `json:"hints,omitempty"`
	Flag          ChallengeImportFlagResp          `json:"flag"`
	Runtime       ChallengeImportRuntimeResp       `json:"runtime"`
	ImageDelivery ChallengeImportImageDeliveryResp `json:"image_delivery"`
	Extensions    ChallengeImportExtensionsResp    `json:"extensions"`
	Topology      *ChallengeImportTopologyResp     `json:"topology,omitempty"`
	PackageFiles  []ChallengePackageFileResp       `json:"package_files,omitempty"`
	Warnings      []string                         `json:"warnings,omitempty"`
	CreatedAt     time.Time                        `json:"created_at"`
}

type ChallengePackageExportResp struct {
	ChallengeID int64     `json:"challenge_id"`
	RevisionID  int64     `json:"revision_id"`
	ArchivePath string    `json:"archive_path"`
	SourceDir   string    `json:"source_dir"`
	FileName    string    `json:"file_name"`
	DownloadURL string    `json:"download_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChallengeImportImageDeliveryResp struct {
	SourceType     string `json:"source_type,omitempty"`
	SuggestedTag   string `json:"suggested_tag,omitempty"`
	TargetImageRef string `json:"target_image_ref,omitempty"`
	BuildStatus    string `json:"build_status,omitempty"`
	Digest         string `json:"digest,omitempty"`
	LastError      string `json:"last_error,omitempty"`
}

type AWDChallengeImportPreviewResp struct {
	ID               string                           `json:"id"`
	FileName         string                           `json:"file_name"`
	Slug             string                           `json:"slug"`
	Title            string                           `json:"title"`
	Category         string                           `json:"category"`
	Difficulty       string                           `json:"difficulty"`
	Description      string                           `json:"description"`
	ServiceType      string                           `json:"service_type"`
	DeploymentMode   string                           `json:"deployment_mode"`
	Version          string                           `json:"version"`
	CheckerType      string                           `json:"checker_type"`
	CheckerConfig    map[string]any                   `json:"checker_config,omitempty"`
	FlagMode         string                           `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any                   `json:"flag_config,omitempty"`
	DefenseEntryMode string                           `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any                   `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any                   `json:"runtime_config,omitempty"`
	ImageDelivery    ChallengeImportImageDeliveryResp `json:"image_delivery"`
	Warnings         []string                         `json:"warnings,omitempty"`
	CreatedAt        time.Time                        `json:"created_at"`
}

type AWDChallengeResp struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Category         string         `json:"category"`
	Difficulty       string         `json:"difficulty"`
	Description      string         `json:"description"`
	ServiceType      string         `json:"service_type"`
	DeploymentMode   string         `json:"deployment_mode"`
	Version          string         `json:"version"`
	Status           string         `json:"status"`
	ReadinessStatus  string         `json:"readiness_status"`
	CheckerType      string         `json:"checker_type,omitempty"`
	CheckerConfig    map[string]any `json:"checker_config,omitempty"`
	FlagMode         string         `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any `json:"flag_config,omitempty"`
	DefenseEntryMode string         `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any `json:"runtime_config,omitempty"`
	CreatedBy        *int64         `json:"created_by,omitempty"`
	LastVerifiedAt   *time.Time     `json:"last_verified_at,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at"`
	CreatedAt        time.Time      `json:"created_at"`
}

type AWDChallengePageResp struct {
	Items []*AWDChallengeResp `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

type AWDChallengeImportCommitResp struct {
	Challenge *AWDChallengeResp `json:"challenge"`
}

type TopologyResourcesResp struct {
	CPUQuota  float64 `json:"cpu_quota,omitempty"`
	MemoryMB  int64   `json:"memory_mb,omitempty"`
	PidsLimit int64   `json:"pids_limit,omitempty"`
}

type TopologyNetworkResp struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	CIDR     string `json:"cidr,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}

type TopologyNodeResp struct {
	Key             string                 `json:"key"`
	Name            string                 `json:"name"`
	ImageID         int64                  `json:"image_id,omitempty"`
	ServicePort     int                    `json:"service_port,omitempty"`
	ServiceProtocol string                 `json:"service_protocol,omitempty"`
	InjectFlag      bool                   `json:"inject_flag,omitempty"`
	Tier            string                 `json:"tier,omitempty"`
	NetworkKeys     []string               `json:"network_keys,omitempty"`
	Env             map[string]string      `json:"env,omitempty"`
	Resources       *TopologyResourcesResp `json:"resources,omitempty"`
}

type TopologyLinkResp struct {
	FromNodeKey string `json:"from_node_key"`
	ToNodeKey   string `json:"to_node_key"`
}

type TopologyTrafficPolicyResp struct {
	SourceNodeKey string `json:"source_node_key"`
	TargetNodeKey string `json:"target_node_key"`
	Action        string `json:"action"`
	Protocol      string `json:"protocol,omitempty"`
	Ports         []int  `json:"ports,omitempty"`
}

type TopologySpecResp struct {
	EntryNodeKey string                      `json:"entry_node_key"`
	Networks     []TopologyNetworkResp       `json:"networks,omitempty"`
	Nodes        []TopologyNodeResp          `json:"nodes"`
	Links        []TopologyLinkResp          `json:"links,omitempty"`
	Policies     []TopologyTrafficPolicyResp `json:"policies,omitempty"`
}

type ChallengePackageFileResp struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type ChallengePackageRevisionResp struct {
	ID                 int64     `json:"id"`
	RevisionNo         int       `json:"revision_no"`
	SourceType         string    `json:"source_type"`
	ParentRevisionID   *int64    `json:"parent_revision_id,omitempty"`
	PackageSlug        string    `json:"package_slug,omitempty"`
	ArchivePath        string    `json:"archive_path,omitempty"`
	SourceDir          string    `json:"source_dir,omitempty"`
	TopologySourcePath string    `json:"topology_source_path,omitempty"`
	CreatedBy          *int64    `json:"created_by,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ChallengeTopologyResp struct {
	ID                   int64                          `json:"id"`
	ChallengeID          int64                          `json:"challenge_id"`
	TemplateID           *int64                         `json:"template_id,omitempty"`
	EntryNodeKey         string                         `json:"entry_node_key"`
	Networks             []TopologyNetworkResp          `json:"networks,omitempty"`
	Nodes                []TopologyNodeResp             `json:"nodes"`
	Links                []TopologyLinkResp             `json:"links,omitempty"`
	Policies             []TopologyTrafficPolicyResp    `json:"policies,omitempty"`
	SourceType           string                         `json:"source_type,omitempty"`
	SourcePath           string                         `json:"source_path,omitempty"`
	SyncStatus           string                         `json:"sync_status,omitempty"`
	PackageRevisionID    *int64                         `json:"package_revision_id,omitempty"`
	LastExportRevisionID *int64                         `json:"last_export_revision_id,omitempty"`
	PackageBaseline      *TopologySpecResp              `json:"package_baseline,omitempty"`
	PackageFiles         []ChallengePackageFileResp     `json:"package_files,omitempty"`
	PackageRevisions     []ChallengePackageRevisionResp `json:"package_revisions,omitempty"`
	CreatedAt            time.Time                      `json:"created_at"`
	UpdatedAt            time.Time                      `json:"updated_at"`
}

type EnvironmentTemplateResp struct {
	ID           int64                       `json:"id"`
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	EntryNodeKey string                      `json:"entry_node_key"`
	Networks     []TopologyNetworkResp       `json:"networks,omitempty"`
	Nodes        []TopologyNodeResp          `json:"nodes"`
	Links        []TopologyLinkResp          `json:"links,omitempty"`
	Policies     []TopologyTrafficPolicyResp `json:"policies,omitempty"`
	UsageCount   int                         `json:"usage_count"`
	CreatedAt    time.Time                   `json:"created_at"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}
