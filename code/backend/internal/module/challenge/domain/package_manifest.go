package domain

type ChallengePackageManifest struct {
	APIVersion string                     `yaml:"api_version"`
	Kind       string                     `yaml:"kind"`
	Meta       ChallengePackageMeta       `yaml:"meta"`
	Content    ChallengePackageContent    `yaml:"content"`
	Flag       ChallengePackageFlag       `yaml:"flag"`
	Hints      []ChallengePackageHint     `yaml:"hints"`
	Runtime    ChallengePackageRuntime    `yaml:"runtime"`
	Extensions ChallengePackageExtensions `yaml:"extensions"`
}

type ChallengePackageMeta struct {
	Slug       string   `yaml:"slug"`
	Title      string   `yaml:"title"`
	Category   string   `yaml:"category"`
	Difficulty string   `yaml:"difficulty"`
	Points     int      `yaml:"points"`
	Mode       string   `yaml:"mode"`
	Tags       []string `yaml:"tags"`
}

type ChallengePackageContent struct {
	Statement   string                       `yaml:"statement"`
	Attachments []ChallengePackageAttachment `yaml:"attachments"`
}

type ChallengePackageAttachment struct {
	Path string `yaml:"path"`
	Name string `yaml:"name"`
}

type ChallengePackageFlag struct {
	Type   string `yaml:"type"`
	Value  string `yaml:"value"`
	Prefix string `yaml:"prefix"`
}

type ChallengePackageHint struct {
	Level   int    `yaml:"level"`
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

type ChallengePackageRuntime struct {
	Type    string                         `yaml:"type"`
	Image   ChallengePackageRuntimeImage   `yaml:"image"`
	Service ChallengePackageRuntimeService `yaml:"service"`
}

type ChallengePackageRuntimeImage struct {
	Ref  string `yaml:"ref"`
	Name string `yaml:"name"`
	Tag  string `yaml:"tag"`
}

type ChallengePackageRuntimeService struct {
	Protocol string `yaml:"protocol"`
	Port     int    `yaml:"port"`
}

type ChallengePackageExtensions struct {
	Topology ChallengePackageTopologyExtension `yaml:"topology"`
	AWD      ChallengePackageAWDExtension      `yaml:"awd"`
}

type ChallengePackageTopologyExtension struct {
	Source  string `yaml:"source"`
	Enabled bool   `yaml:"enabled"`
}

type ChallengePackageTopologyManifest struct {
	APIVersion   string                            `yaml:"api_version"`
	Kind         string                            `yaml:"kind"`
	EntryNodeKey string                            `yaml:"entry_node_key"`
	Networks     []ChallengePackageTopologyNetwork `yaml:"networks"`
	Nodes        []ChallengePackageTopologyNode    `yaml:"nodes"`
	Links        []ChallengePackageTopologyLink    `yaml:"links"`
	Policies     []ChallengePackageTopologyPolicy  `yaml:"policies"`
}

type ChallengePackageTopologyNetwork struct {
	Key      string `yaml:"key"`
	Name     string `yaml:"name"`
	CIDR     string `yaml:"cidr"`
	Internal bool   `yaml:"internal"`
}

type ChallengePackageTopologyNode struct {
	Key         string                             `yaml:"key"`
	Name        string                             `yaml:"name"`
	Tier        string                             `yaml:"tier"`
	Image       ChallengePackageTopologyNodeImage  `yaml:"image"`
	ServicePort int                                `yaml:"service_port"`
	InjectFlag  bool                               `yaml:"inject_flag"`
	NetworkKeys []string                           `yaml:"network_keys"`
	Env         map[string]string                  `yaml:"env"`
	Resources   *ChallengePackageTopologyResources `yaml:"resources"`
}

type ChallengePackageTopologyNodeImage struct {
	Ref        string `yaml:"ref"`
	Dockerfile string `yaml:"dockerfile"`
	Context    string `yaml:"context"`
}

type ChallengePackageTopologyResources struct {
	CPUQuota  float64 `yaml:"cpu_quota"`
	MemoryMB  int64   `yaml:"memory_mb"`
	PidsLimit int64   `yaml:"pids_limit"`
}

type ChallengePackageTopologyLink struct {
	FromNodeKey string `yaml:"from_node_key"`
	ToNodeKey   string `yaml:"to_node_key"`
}

type ChallengePackageTopologyPolicy struct {
	SourceNodeKey string `yaml:"source_node_key"`
	TargetNodeKey string `yaml:"target_node_key"`
	Action        string `yaml:"action"`
	Protocol      string `yaml:"protocol"`
	Ports         []int  `yaml:"ports"`
}

type ChallengePackageAWDExtension struct {
	ServiceType    string                          `yaml:"service_type"`
	DeploymentMode string                          `yaml:"deployment_mode"`
	Version        string                          `yaml:"version"`
	Checker        ChallengePackageAWDChecker      `yaml:"checker"`
	FlagPolicy     ChallengePackageAWDFlagPolicy   `yaml:"flag_policy"`
	DefenseEntry   ChallengePackageAWDDefenseEntry `yaml:"defense_entry"`
	AccessConfig   map[string]any                  `yaml:"access_config"`
	RuntimeConfig  map[string]any                  `yaml:"runtime_config"`
}

type ChallengePackageAWDChecker struct {
	Type   string         `yaml:"type"`
	Config map[string]any `yaml:"config"`
}

type ChallengePackageAWDFlagPolicy struct {
	Mode   string         `yaml:"mode"`
	Config map[string]any `yaml:"config"`
}

type ChallengePackageAWDDefenseEntry struct {
	Mode string `yaml:"mode"`
}

type ParsedChallengePackage struct {
	Manifest        ChallengePackageManifest
	ManifestRaw     string
	RootDir         string
	Slug            string
	Title           string
	Description     string
	Category        string
	Difficulty      string
	Points          int
	FlagType        string
	FlagValue       string
	FlagPrefix      string
	RuntimeImageRef string
	RuntimeProtocol string
	RuntimePort     int
	Attachments     []ParsedChallengePackageAttachment
	Hints           []ParsedChallengePackageHint
	Topology        *ParsedChallengePackageTopology
	PackageFiles    []ParsedChallengePackageFile
	Warnings        []string
}

type ParsedChallengePackageAttachment struct {
	Path         string
	Name         string
	AbsolutePath string
}

type ParsedChallengePackageHint struct {
	Level   int
	Title   string
	Content string
}

type ParsedChallengePackageTopology struct {
	Source       string
	Raw          string
	EntryNodeKey string
	Networks     []ChallengePackageTopologyNetwork
	Nodes        []ChallengePackageTopologyNode
	Links        []ChallengePackageTopologyLink
	Policies     []ChallengePackageTopologyPolicy
}

type ParsedChallengePackageFile struct {
	Path string
	Size int64
}

type ParsedAWDChallengePackage struct {
	Manifest         ChallengePackageManifest
	RootDir          string
	Slug             string
	Title            string
	Description      string
	Category         string
	Difficulty       string
	SuggestedPoints  int
	RuntimeImageRef  string
	ServiceType      string
	DeploymentMode   string
	Version          string
	CheckerType      string
	CheckerConfig    map[string]any
	CheckerEntryPath string
	CheckerEntryAbs  string
	CheckerFiles     []ParsedAWDCheckerFile
	FlagMode         string
	FlagConfig       map[string]any
	DefenseEntryMode string
	AccessConfig     map[string]any
	RuntimeConfig    map[string]any
	Warnings         []string
}

type ParsedAWDCheckerFile struct {
	Path string
	Abs  string
}
