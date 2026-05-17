package commands

import challengecontracts "ctf-platform/internal/module/challenge/contracts"

type SaveChallengeTopologyInput struct {
	TemplateID   *int64
	EntryNodeKey string
	Networks     []challengecontracts.TopologyNetworkReq
	Nodes        []challengecontracts.TopologyNodeReq
	Links        []challengecontracts.TopologyLinkReq
	Policies     []challengecontracts.TopologyTrafficPolicyReq
}

type UpsertEnvironmentTemplateInput struct {
	Name         string
	Description  string
	EntryNodeKey string
	Networks     []challengecontracts.TopologyNetworkReq
	Nodes        []challengecontracts.TopologyNodeReq
	Links        []challengecontracts.TopologyLinkReq
	Policies     []challengecontracts.TopologyTrafficPolicyReq
}
