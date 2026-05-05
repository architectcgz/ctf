package dto

import "time"

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

type ChallengeImportImageDeliveryResp struct {
	SourceType     string `json:"source_type,omitempty"`
	SuggestedTag   string `json:"suggested_tag,omitempty"`
	TargetImageRef string `json:"target_image_ref,omitempty"`
	BuildStatus    string `json:"build_status,omitempty"`
	Digest         string `json:"digest,omitempty"`
	LastError      string `json:"last_error,omitempty"`
}

type ChallengeImportTopologyExtensionResp struct {
	Source  string `json:"source,omitempty"`
	Enabled bool   `json:"enabled"`
}

type ChallengePackageFileResp struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
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

type ChallengeImportCommitResp struct {
	Challenge *ChallengeResp `json:"challenge"`
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
