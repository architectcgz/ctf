package domain

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen .

// goverter:converter
// goverter:output:file ./zz_generated.goverter.go
// goverter:extend FormatImageSize
// goverter:skipCopySameType
type ChallengeResponseMapper interface {
	// goverter:ignore Hints
	ToChallengeResp(source *model.Challenge) *dto.ChallengeResp
	ToChallengeHintAdminResp(source *model.ChallengeHint) *dto.ChallengeHintAdminResp
	// goverter:map ServiceType ServiceType | AWDServiceTypeToString
	// goverter:map DeploymentMode DeploymentMode | AWDDeploymentModeToString
	// goverter:map Status Status | AWDChallengeStatusToString
	// goverter:map ReadinessStatus ReadinessStatus | AWDReadinessStatusToString
	// goverter:ignore CheckerType
	// goverter:ignore CheckerConfig
	// goverter:ignore FlagMode
	// goverter:ignore FlagConfig
	// goverter:ignore DefenseEntryMode
	// goverter:ignore AccessConfig
	// goverter:ignore RuntimeConfig
	ToAWDChallengeResp(source *model.AWDChallenge) *dto.AWDChallengeResp
	// goverter:ignore SizeFormatted
	ToImageResp(source *model.Image) *dto.ImageResp
	ToTagResp(source *model.Tag) *dto.TagResp
	ToAdminChallengeWriteupResp(source *model.ChallengeWriteup) *dto.AdminChallengeWriteupResp
	ToSubmissionWriteupResp(source *model.SubmissionWriteup) *dto.SubmissionWriteupResp
	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore StudentNo
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore ContentPreview
	ToTeacherSubmissionWriteupItemRespBase(source *model.SubmissionWriteup) *dto.TeacherSubmissionWriteupItemResp
	// goverter:ignore AuthorName
	// goverter:ignore ContentPreview
	ToCommunityChallengeSolutionRespBase(source *model.SubmissionWriteup) *dto.CommunityChallengeSolutionResp
	ToChallengePackageRevisionResp(source model.ChallengePackageRevision) dto.ChallengePackageRevisionResp
	// goverter:ignore ID
	ToRecommendedChallengeSolutionRespBase(source challengeports.RecommendedSolutionRecord) *dto.RecommendedChallengeSolutionResp
	// goverter:ignore ImageRef
	ToChallengeImportTopologyNodeRespBase(source ChallengePackageTopologyNode) dto.ChallengeImportTopologyNodeResp
	ToChallengePackageFileResp(source ParsedChallengePackageFile) dto.ChallengePackageFileResp
	// goverter:ignore EntryNodeKey
	// goverter:ignore Networks
	// goverter:ignore Nodes
	// goverter:ignore Links
	// goverter:ignore Policies
	ToTopologySpecRespBase(source model.TopologySpec) *dto.TopologySpecResp
	// goverter:ignore Networks
	// goverter:ignore Nodes
	// goverter:ignore Links
	// goverter:ignore Policies
	ToChallengeImportTopologyRespBase(source ParsedChallengePackageTopology) *dto.ChallengeImportTopologyResp
	ToImportedTopologyNetworkList(source []ChallengePackageTopologyNetwork) []model.TopologyNetwork
	ToImportedTopologyLinkList(source []ChallengePackageTopologyLink) []model.TopologyLink
	ToImportedTopologyPolicyList(source []ChallengePackageTopologyPolicy) []model.TopologyTrafficPolicy
	// goverter:ignore Networks
	// goverter:ignore Nodes
	// goverter:ignore Links
	// goverter:ignore Policies
	// goverter:ignore PackageBaseline
	// goverter:ignore PackageFiles
	// goverter:ignore PackageRevisions
	ToChallengeTopologyRespBase(source *model.ChallengeTopology) *dto.ChallengeTopologyResp
	// goverter:ignore Networks
	// goverter:ignore Nodes
	// goverter:ignore Links
	// goverter:ignore Policies
	ToEnvironmentTemplateRespBase(source *model.EnvironmentTemplate) *dto.EnvironmentTemplateResp
	ToTopologyNetworkRespList(source []model.TopologyNetwork) []dto.TopologyNetworkResp
	ToTopologyNodeRespList(source []model.TopologyNode) []dto.TopologyNodeResp
	ToTopologyLinkRespList(source []model.TopologyLink) []dto.TopologyLinkResp
	ToTopologyTrafficPolicyRespList(source []model.TopologyTrafficPolicy) []dto.TopologyTrafficPolicyResp
}

var challengeResponseMapperInst ChallengeResponseMapper

func AWDServiceTypeToString(value model.AWDServiceType) string {
	return string(value)
}

func AWDDeploymentModeToString(value model.AWDDeploymentMode) string {
	return string(value)
}

func AWDChallengeStatusToString(value model.AWDChallengeStatus) string {
	return string(value)
}

func AWDReadinessStatusToString(value model.AWDReadinessStatus) string {
	return string(value)
}
