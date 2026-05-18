package domain

import (
	"time"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :domain
type ChallengeResponseMapper interface {
	// goverter:ignore Hints
	ToChallengeRespBase(source model.Challenge) challengecontracts.ChallengeResp
	ToChallengeRespBasePtr(source *model.Challenge) *challengecontracts.ChallengeResp

	// goverter:ignore Networks
	// goverter:ignore Nodes
	// goverter:ignore Links
	// goverter:ignore Policies
	// goverter:ignore PackageBaseline
	// goverter:ignore PackageFiles
	// goverter:ignore PackageRevisions
	ToChallengeTopologyRespBase(source model.ChallengeTopology) challengecontracts.ChallengeTopologyResp
	ToChallengeTopologyRespBasePtr(source *model.ChallengeTopology) *challengecontracts.ChallengeTopologyResp

	// goverter:ignore Networks
	// goverter:ignore Nodes
	// goverter:ignore Links
	// goverter:ignore Policies
	ToEnvironmentTemplateRespBase(source model.EnvironmentTemplate) challengecontracts.EnvironmentTemplateResp
	ToEnvironmentTemplateRespBasePtr(source *model.EnvironmentTemplate) *challengecontracts.EnvironmentTemplateResp

	// goverter:ignore CheckerConfig FlagConfig AccessConfig RuntimeConfig
	ToAWDChallengeRespBase(source challengeentity.AWDChallenge) challengecontracts.AWDChallengeResp
	ToAWDChallengeRespBasePtr(source *challengeentity.AWDChallenge) *challengecontracts.AWDChallengeResp

	ToChallengeHintAdminResp(source challengeentity.ChallengeHint) challengecontracts.ChallengeHintAdminResp
	ToChallengeHintAdminRespPtr(source *challengeentity.ChallengeHint) *challengecontracts.ChallengeHintAdminResp

	// goverter:ignore SizeFormatted
	ToImageRespBase(source model.Image) challengecontracts.ImageResp
	ToImageRespBasePtr(source *model.Image) *challengecontracts.ImageResp

	ToTagResp(source model.Tag) challengecontracts.TagResp
	ToTagRespPtr(source *model.Tag) *challengecontracts.TagResp
	ToAdminChallengeWriteupResp(source model.ChallengeWriteup) challengecontracts.AdminChallengeWriteupResp
	ToAdminChallengeWriteupRespPtr(source *model.ChallengeWriteup) *challengecontracts.AdminChallengeWriteupResp
	ToSubmissionWriteupResp(source challengeentity.SubmissionWriteup) challengecontracts.SubmissionWriteupResp
	ToSubmissionWriteupRespPtr(source *challengeentity.SubmissionWriteup) *challengecontracts.SubmissionWriteupResp

	// goverter:map Submission.ID ID
	// goverter:map Submission.UserID UserID
	// goverter:map Submission.ChallengeID ChallengeID
	// goverter:map Submission.Title Title
	// goverter:map Submission.SubmissionStatus SubmissionStatus
	// goverter:map Submission.VisibilityStatus VisibilityStatus
	// goverter:map Submission.IsRecommended IsRecommended
	// goverter:map Submission.PublishedAt PublishedAt
	// goverter:map Submission.UpdatedAt UpdatedAt
	// goverter:ignore ContentPreview
	ToTeacherSubmissionWriteupItemRespBase(source challengeports.TeacherSubmissionWriteupRecord) challengecontracts.TeacherSubmissionWriteupItemResp
	ToTeacherSubmissionWriteupItemRespBasePtr(source *challengeports.TeacherSubmissionWriteupRecord) *challengecontracts.TeacherSubmissionWriteupItemResp

	// goverter:map Submission SubmissionWriteupResp
	ToTeacherSubmissionWriteupDetailResp(source challengeports.TeacherSubmissionWriteupRecord) challengecontracts.TeacherSubmissionWriteupDetailResp
	ToTeacherSubmissionWriteupDetailRespPtr(source challengeports.TeacherSubmissionWriteupRecord) *challengecontracts.TeacherSubmissionWriteupDetailResp

	// goverter:ignore ID
	ToRecommendedChallengeSolutionRespBase(source challengeports.RecommendedSolutionRecord) challengecontracts.RecommendedChallengeSolutionResp
	ToRecommendedChallengeSolutionRespBasePtr(source *challengeports.RecommendedSolutionRecord) *challengecontracts.RecommendedChallengeSolutionResp

	ToChallengePackageRevisionResp(source challengeentity.ChallengePackageRevision) challengecontracts.ChallengePackageRevisionResp
	ToChallengePackageRevisionRespPtr(source *challengeentity.ChallengePackageRevision) *challengecontracts.ChallengePackageRevisionResp

	ToChallengePackageFileResp(source ParsedChallengePackageFile) challengecontracts.ChallengePackageFileResp
	ToChallengePackageFileResps(source []ParsedChallengePackageFile) []challengecontracts.ChallengePackageFileResp

	ToTopologyNetworkResp(source model.TopologyNetwork) challengecontracts.TopologyNetworkResp
	ToTopologyNetworkResps(source []model.TopologyNetwork) []challengecontracts.TopologyNetworkResp
	ToTopologyNodeResp(source model.TopologyNode) challengecontracts.TopologyNodeResp
	ToTopologyNodeResps(source []model.TopologyNode) []challengecontracts.TopologyNodeResp
	ToTopologyLinkResp(source model.TopologyLink) challengecontracts.TopologyLinkResp
	ToTopologyLinkResps(source []model.TopologyLink) []challengecontracts.TopologyLinkResp
	ToTopologyTrafficPolicyResp(source model.TopologyTrafficPolicy) challengecontracts.TopologyTrafficPolicyResp
	ToTopologyTrafficPolicyResps(source []model.TopologyTrafficPolicy) []challengecontracts.TopologyTrafficPolicyResp

	ToImportedTopologyNetwork(source ChallengePackageTopologyNetwork) model.TopologyNetwork
	ToImportedTopologyNetworks(source []ChallengePackageTopologyNetwork) []model.TopologyNetwork
	ToImportedTopologyLink(source ChallengePackageTopologyLink) model.TopologyLink
	ToImportedTopologyLinks(source []ChallengePackageTopologyLink) []model.TopologyLink
	ToImportedTopologyPolicy(source ChallengePackageTopologyPolicy) model.TopologyTrafficPolicy
	ToImportedTopologyPolicies(source []ChallengePackageTopologyPolicy) []model.TopologyTrafficPolicy

	// goverter:map Image.Ref ImageRef
	ToChallengeImportTopologyNodeRespBase(source ChallengePackageTopologyNode) challengecontracts.ChallengeImportTopologyNodeResp
	ToChallengeImportTopologyNodeRespBases(source []ChallengePackageTopologyNode) []challengecontracts.ChallengeImportTopologyNodeResp

	// goverter:map Submission.ID ID
	// goverter:map Submission.ChallengeID ChallengeID
	// goverter:map Submission.UserID UserID
	// goverter:map Submission.Title Title
	// goverter:map Submission.Content Content
	// goverter:map Submission.SubmissionStatus SubmissionStatus
	// goverter:map Submission.VisibilityStatus VisibilityStatus
	// goverter:map Submission.IsRecommended IsRecommended
	// goverter:map Submission.PublishedAt PublishedAt
	// goverter:map Submission.UpdatedAt UpdatedAt
	// goverter:ignore ContentPreview
	ToCommunityChallengeSolutionRespBase(source challengeports.CommunitySolutionRecord) challengecontracts.CommunityChallengeSolutionResp
	ToCommunityChallengeSolutionRespBasePtr(source *challengeports.CommunitySolutionRecord) *challengecontracts.CommunityChallengeSolutionResp
}

var challengeResponseMapperInst ChallengeResponseMapper

func ResponseMapper() ChallengeResponseMapper {
	return challengeResponseMapperInst
}

func CopyTime(value time.Time) time.Time {
	return value
}

func CopyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
