package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/practice/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

func (s *Service) createContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, topology *model.ChallengeTopology, flag string) error {
	if topology == nil {
		return s.createSingleContainer(ctx, instance, chal, flag)
	}

	awdWorkspacePlan, err := s.prepareAWDDefenseWorkspacePlan(ctx, instance, chal)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	if awdWorkspacePlan != nil && awdWorkspacePlan.createWorkspace {
		if err := s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusProvisioning, ""); err != nil {
			return errcode.ErrContainerCreateFailed.WithCause(err)
		}
	}

	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	request, err := s.buildTopologyCreateRequest(ctx, instance.HostPort, isAWDInstance(instance), chal, topology.EntryNodeKey, spec, flag)
	if err != nil {
		return err
	}
	applyAWDStableNetworkToTopologyRequest(instance, chal, request)
	if awdWorkspacePlan != nil {
		applyAWDDefenseWorkspaceRuntimeMounts(request, awdWorkspacePlan.runtimeMounts)
		applyAWDCheckerTokenToTopologyRequest(request, awdWorkspacePlan.checkerTokenEnv, awdWorkspacePlan.checkerToken)
	}
	result, err := s.runtimeService.CreateTopology(ctx, request)
	if err != nil {
		if awdWorkspacePlan != nil && awdWorkspacePlan.createWorkspace {
			_ = s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusFailed, "")
		}
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	if awdWorkspacePlan != nil {
		workspaceContainerID := awdWorkspacePlan.workspaceContainerID
		if awdWorkspacePlan.createWorkspace {
			workspaceContainerID, err = s.createAWDDefenseWorkspaceCompanion(ctx, instance, awdWorkspacePlan)
			if err != nil {
				_ = s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusFailed, "")
				return errcode.ErrContainerCreateFailed.WithCause(err)
			}
		}
		if err := s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusRunning, workspaceContainerID); err != nil {
			return errcode.ErrContainerCreateFailed.WithCause(err)
		}
	}
	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(result.RuntimeDetails)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	instance.ContainerID = result.PrimaryContainerID
	instance.NetworkID = result.NetworkID
	instance.RuntimeDetails = runtimeDetails
	instance.AccessURL = result.AccessURL
	return nil
}

func (s *Service) createSingleContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, flag string) error {
	imageItem, err := s.imageRepo.FindByID(ctx, chal.ImageID)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}
	if imageItem.Status != model.ImageStatusAvailable {
		return errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("image %d is not available", imageItem.ID))
	}

	env := map[string]string{
		"FLAG": flag,
	}

	imageRef := model.BuildRuntimeImageRef(imageItem)
	targetProtocol := normalizeChallengeTargetProtocol(chal.TargetProtocol)
	if isAWDInstance(instance) || targetProtocol == model.ChallengeTargetProtocolTCP || chal.TargetPort > 0 {
		awdWorkspacePlan, err := s.prepareAWDDefenseWorkspacePlan(ctx, instance, chal)
		if err != nil {
			return errcode.ErrContainerCreateFailed.WithCause(err)
		}
		if awdWorkspacePlan != nil && awdWorkspacePlan.createWorkspace {
			if err := s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusProvisioning, ""); err != nil {
				return errcode.ErrContainerCreateFailed.WithCause(err)
			}
		}
		runtimeMounts := []model.ContainerMount(nil)
		if awdWorkspacePlan != nil {
			runtimeMounts = append(runtimeMounts, awdWorkspacePlan.runtimeMounts...)
			if awdWorkspacePlan.checkerTokenEnv != "" {
				env[awdWorkspacePlan.checkerTokenEnv] = awdWorkspacePlan.checkerToken
			}
		}

		networks := []practiceports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		}
		nodeAliases := []string(nil)
		if isAWDInstance(instance) {
			networks[0].Name = buildAWDContestNetworkName(instance)
			networks[0].Shared = true
			nodeAliases = []string{buildAWDServiceAlias(instance)}
		}
		result, err := s.runtimeService.CreateTopology(ctx, &practiceports.TopologyCreateRequest{
			ReservedHostPort:           instance.HostPort,
			DisableEntryPortPublishing: isAWDInstance(instance),
			ContainerName:              buildRuntimeContainerName(chal, instance),
			Networks:                   networks,
			Nodes: []practiceports.TopologyCreateNode{
				{
					Key:             "default",
					Image:           imageRef,
					Env:             env,
					ServicePort:     chal.TargetPort,
					ServiceProtocol: targetProtocol,
					IsEntryPoint:    true,
					NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
					NetworkAliases:  nodeAliases,
					Mounts:          runtimeMounts,
				},
			},
		})
		if err != nil {
			if awdWorkspacePlan != nil && awdWorkspacePlan.createWorkspace {
				_ = s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusFailed, "")
			}
			return errcode.ErrContainerCreateFailed.WithCause(err)
		}
		if awdWorkspacePlan != nil {
			workspaceContainerID := awdWorkspacePlan.workspaceContainerID
			if awdWorkspacePlan.createWorkspace {
				workspaceContainerID, err = s.createAWDDefenseWorkspaceCompanion(ctx, instance, awdWorkspacePlan)
				if err != nil {
					_ = s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusFailed, "")
					return errcode.ErrContainerCreateFailed.WithCause(err)
				}
			}
			if err := s.persistAWDDefenseWorkspaceState(ctx, awdWorkspacePlan, instance.ID, model.AWDDefenseWorkspaceStatusRunning, workspaceContainerID); err != nil {
				return errcode.ErrContainerCreateFailed.WithCause(err)
			}
		}
		runtimeDetails, err := model.EncodeInstanceRuntimeDetails(result.RuntimeDetails)
		if err != nil {
			return errcode.ErrContainerCreateFailed.WithCause(err)
		}
		instance.ContainerID = result.PrimaryContainerID
		instance.NetworkID = result.NetworkID
		instance.RuntimeDetails = runtimeDetails
		instance.AccessURL = result.AccessURL
		return nil
	}

	containerID, networkID, hostPort, servicePort, err := s.runtimeService.CreateContainer(ctx, imageRef, env, instance.HostPort)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Containers: []model.InstanceRuntimeContainer{
			{
				NodeKey:         "default",
				ContainerID:     containerID,
				ServicePort:     servicePort,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				HostPort:        hostPort,
				IsEntryPoint:    true,
			},
		},
	})
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	instance.ContainerID = containerID
	instance.NetworkID = networkID
	instance.RuntimeDetails = runtimeDetails
	instance.AccessURL = fmt.Sprintf("http://%s:%d", s.config.Container.PublicHost, hostPort)
	return nil
}

func normalizeChallengeTargetProtocol(protocol string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case model.ChallengeTargetProtocolTCP:
		return model.ChallengeTargetProtocolTCP
	default:
		return model.ChallengeTargetProtocolHTTP
	}
}

func (s *Service) buildTopologyCreateRequest(
	ctx context.Context,
	reservedHostPort int,
	disableEntryPortPublishing bool,
	chal *model.Challenge,
	entryNodeKey string,
	spec model.TopologySpec,
	flag string,
) (*practiceports.TopologyCreateRequest, error) {
	if len(spec.Nodes) == 0 {
		return nil, errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("challenge topology has no nodes"))
	}
	if chal != nil && chal.InstanceSharing == model.InstanceSharingShared {
		for _, node := range spec.Nodes {
			if node.InjectFlag {
				return nil, errcode.ErrInvalidParams.WithCause(errors.New("共享实例策略不支持带 Flag 注入的拓扑"))
			}
		}
	}

	defaultImageRef, err := s.resolveAvailableImageRef(ctx, chal.ImageID)
	if err != nil {
		return nil, err
	}

	request := &practiceports.TopologyCreateRequest{
		ReservedHostPort:           reservedHostPort,
		DisableEntryPortPublishing: disableEntryPortPublishing,
		Networks:                   make([]practiceports.TopologyCreateNetwork, 0),
		Nodes:                      make([]practiceports.TopologyCreateNode, 0, len(spec.Nodes)),
		Policies:                   append([]model.TopologyTrafficPolicy(nil), spec.Policies...),
	}
	runtimePlan := domain.BuildRuntimeTopologyPlan(spec)
	request.Networks = append(request.Networks, runtimePlan.Networks...)
	for _, node := range spec.Nodes {
		imageRef := defaultImageRef
		if node.ImageID > 0 {
			imageRef, err = s.resolveAvailableImageRef(ctx, node.ImageID)
			if err != nil {
				return nil, err
			}
		}

		env := make(map[string]string, len(node.Env)+1)
		for key, value := range node.Env {
			env[key] = value
		}
		if node.InjectFlag {
			env["FLAG"] = flag
		}

		var resources *model.ResourceLimits
		if node.Resources != nil {
			resources = &model.ResourceLimits{
				CPUQuota:  node.Resources.CPUQuota,
				Memory:    node.Resources.MemoryMB * 1024 * 1024,
				PidsLimit: node.Resources.PidsLimit,
			}
		}

		request.Nodes = append(request.Nodes, practiceports.TopologyCreateNode{
			Key:             node.Key,
			Image:           imageRef,
			Env:             env,
			ServicePort:     node.ServicePort,
			ServiceProtocol: normalizeChallengeTargetProtocol(node.ServiceProtocol),
			IsEntryPoint:    node.Key == entryNodeKey,
			NetworkKeys:     append([]string(nil), runtimePlan.NodeNetworkKeys[node.Key]...),
			Resources:       resources,
		})
	}

	return request, nil
}

func (s *Service) resolveAvailableImageRef(ctx context.Context, imageID int64) (string, error) {
	imageItem, err := s.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		return "", errcode.ErrContainerCreateFailed.WithCause(err)
	}
	if imageItem.Status != model.ImageStatusAvailable {
		return "", errcode.ErrContainerCreateFailed.WithCause(fmt.Errorf("image %d is not available", imageItem.ID))
	}
	return model.BuildRuntimeImageRef(imageItem), nil
}

func applyAWDCheckerTokenToTopologyRequest(req *practiceports.TopologyCreateRequest, checkerTokenEnv, checkerToken string) {
	if req == nil || strings.TrimSpace(checkerTokenEnv) == "" || strings.TrimSpace(checkerToken) == "" {
		return
	}
	for index := range req.Nodes {
		if req.Nodes[index].Env == nil {
			continue
		}
		if _, ok := req.Nodes[index].Env["FLAG"]; !ok {
			continue
		}
		req.Nodes[index].Env[checkerTokenEnv] = checkerToken
	}
}
