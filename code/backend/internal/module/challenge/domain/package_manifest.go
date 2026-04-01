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
	Level      int    `yaml:"level"`
	Title      string `yaml:"title"`
	CostPoints int    `yaml:"cost_points"`
	Content    string `yaml:"content"`
}

type ChallengePackageRuntime struct {
	Type  string                       `yaml:"type"`
	Image ChallengePackageRuntimeImage `yaml:"image"`
}

type ChallengePackageRuntimeImage struct {
	Ref  string `yaml:"ref"`
	Name string `yaml:"name"`
	Tag  string `yaml:"tag"`
}

type ChallengePackageExtensions struct {
	Topology ChallengePackageTopologyExtension `yaml:"topology"`
}

type ChallengePackageTopologyExtension struct {
	Source  string `yaml:"source"`
	Enabled bool   `yaml:"enabled"`
}

type ParsedChallengePackage struct {
	Manifest        ChallengePackageManifest
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
	Attachments     []ParsedChallengePackageAttachment
	Hints           []ParsedChallengePackageHint
	Warnings        []string
}

type ParsedChallengePackageAttachment struct {
	Path         string
	Name         string
	AbsolutePath string
}

type ParsedChallengePackageHint struct {
	Level      int
	Title      string
	CostPoints int
	Content    string
}
