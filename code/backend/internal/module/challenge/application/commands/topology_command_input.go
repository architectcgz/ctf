package commands

import "ctf-platform/internal/dto"

type SaveChallengeTopologyInput struct {
	TemplateID   *int64
	EntryNodeKey string
	Networks     []dto.TopologyNetworkReq
	Nodes        []dto.TopologyNodeReq
	Links        []dto.TopologyLinkReq
	Policies     []dto.TopologyTrafficPolicyReq
}

type UpsertEnvironmentTemplateInput struct {
	Name         string
	Description  string
	EntryNodeKey string
	Networks     []dto.TopologyNetworkReq
	Nodes        []dto.TopologyNodeReq
	Links        []dto.TopologyLinkReq
	Policies     []dto.TopologyTrafficPolicyReq
}
