package challengeselfcheck

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	"ctf-platform/internal/platform/randomstring"
	crypto "ctf-platform/internal/shared/flagcrypto"
)

type challengeWriteLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*challengeports.ChallengeWriteModel, error)
}

type Config struct {
	RuntimeCreateTimeout time.Duration
	FlagGlobalSecret     string
}

type ChallengeSelfCheckService struct {
	repo         challengeWriteLookupRepository
	imageRepo    challengeports.ImageQueryRepository
	topologyRepo challengeports.ChallengeTopologyReadRepository
	runtimeProbe challengeports.ChallengeRuntimeProbe
	selfCheckCfg Config
	logger       *zap.Logger
}

func NewChallengeSelfCheckService(
	repo challengeWriteLookupRepository,
	imageRepo challengeports.ImageQueryRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
	cfg Config,
	logger *zap.Logger,
) *ChallengeSelfCheckService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg.RuntimeCreateTimeout <= 0 {
		cfg.RuntimeCreateTimeout = 60 * time.Second
	}
	return &ChallengeSelfCheckService{
		repo:         repo,
		imageRepo:    imageRepo,
		topologyRepo: topologyRepo,
		runtimeProbe: runtimeProbe,
		selfCheckCfg: cfg,
		logger:       logger,
	}
}

type challengeSelfCheckRuntimeInput struct {
	defaultImageRef string
	nodeImageRefs   map[int64]string
	entryNodeKey    string
	topologySpec    challengecontracts.TopologySpec
	useTopology     bool
	skipRuntime     bool
}

func (s *ChallengeSelfCheckService) SelfCheckChallenge(ctx context.Context, id int64) (*challengecontracts.ChallengeSelfCheckResp, error) {
	challengeWriteModel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, err
	}
	challenge := challengeWriteModel

	resp := &challengecontracts.ChallengeSelfCheckResp{
		ChallengeID: challenge.ID,
	}

	resp.Precheck.StartedAt = time.Now().UTC()
	input, precheckPassed, err := s.runPrecheck(ctx, challenge, &resp.Precheck.Steps)
	resp.Precheck.EndedAt = time.Now().UTC()
	if err != nil {
		return nil, err
	}
	resp.Precheck.Passed = precheckPassed

	resp.Runtime.StartedAt = time.Now().UTC()
	if !resp.Precheck.Passed {
		resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "runtime_startup",
			Passed:  false,
			Message: "预检未通过，已跳过真实拉起",
		})
		resp.Runtime.EndedAt = time.Now().UTC()
		return resp, nil
	}
	if input.skipRuntime {
		resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "runtime_startup",
			Passed:  true,
			Message: "当前题目无需运行时，已跳过真实拉起",
		})
		resp.Runtime.EndedAt = time.Now().UTC()
		resp.Runtime.Passed = true
		return resp, nil
	}
	if s.runtimeProbe == nil {
		resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "runtime_startup",
			Passed:  false,
			Message: "运行时自测能力未配置",
		})
		resp.Runtime.EndedAt = time.Now().UTC()
		return resp, nil
	}

	createCtx, cancel := context.WithTimeout(ctx, s.selfCheckCfg.RuntimeCreateTimeout)
	defer cancel()

	runtimePassed := true
	flag, flagErr := s.buildRuntimeFlag(challenge)
	if flagErr != nil {
		runtimePassed = false
		resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "runtime_flag",
			Passed:  false,
			Message: fmt.Sprintf("生成运行时 Flag 失败: %v", flagErr),
		})
	} else {
		resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "runtime_flag",
			Passed:  true,
			Message: "运行时 Flag 已准备",
		})
	}

	var (
		runtimeDetails runtimecontracts.InstanceRuntimeDetails
		accessURL      string
	)
	if runtimePassed {
		if input.useTopology {
			req, buildErr := s.buildTopologyRuntimeRequest(input, flag)
			if buildErr != nil {
				runtimePassed = false
				resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
					Name:    "runtime_startup",
					Passed:  false,
					Message: fmt.Sprintf("构建拓扑启动请求失败: %v", buildErr),
				})
			} else {
				result, startupErr := s.runtimeProbe.CreateTopology(createCtx, req)
				if startupErr != nil {
					runtimePassed = false
					resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
						Name:    "runtime_startup",
						Passed:  false,
						Message: fmt.Sprintf("拓扑拉起失败: %v", startupErr),
					})
				} else {
					accessURL = result.AccessURL
					runtimeDetails = result.RuntimeDetails
					resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
						Name:    "runtime_startup",
						Passed:  true,
						Message: "拓扑实例拉起成功",
					})
				}
			}
		} else {
			startupAccessURL, details, startupErr := s.runtimeProbe.CreateContainer(createCtx, input.defaultImageRef, map[string]string{
				"FLAG": flag,
			})
			if startupErr != nil {
				runtimePassed = false
				resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
					Name:    "runtime_startup",
					Passed:  false,
					Message: fmt.Sprintf("单容器拉起失败: %v", startupErr),
				})
			} else {
				accessURL = startupAccessURL
				runtimeDetails = details
				resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
					Name:    "runtime_startup",
					Passed:  true,
					Message: "单容器实例拉起成功",
				})
			}
		}
	}

	if runtimePassed {
		if cleanupErr := s.runtimeProbe.CleanupRuntimeDetails(ctx, runtimeDetails); cleanupErr != nil {
			runtimePassed = false
			resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
				Name:    "runtime_cleanup",
				Passed:  false,
				Message: fmt.Sprintf("运行时资源清理失败: %v", cleanupErr),
			})
		} else {
			resp.Runtime.Steps = append(resp.Runtime.Steps, challengecontracts.ChallengeSelfCheckStepResp{
				Name:    "runtime_cleanup",
				Passed:  true,
				Message: "运行时资源已清理",
			})
		}
	}

	resp.Runtime.EndedAt = time.Now().UTC()
	resp.Runtime.Passed = runtimePassed
	resp.Runtime.AccessURL = accessURL
	resp.Runtime.ContainerCount = len(runtimeDetails.Containers)
	resp.Runtime.NetworkCount = len(runtimeDetails.Networks)
	return resp, nil
}

func (s *ChallengeSelfCheckService) runPrecheck(ctx context.Context, challenge *challengeports.ChallengeWriteModel, steps *[]challengecontracts.ChallengeSelfCheckStepResp) (challengeSelfCheckRuntimeInput, bool, error) {
	input := challengeSelfCheckRuntimeInput{
		nodeImageRefs: make(map[int64]string),
	}
	passed := true

	flagOK, flagMessage := s.validateFlagConfig(challenge)
	*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
		Name:    "flag_config",
		Passed:  flagOK,
		Message: flagMessage,
	})
	if !flagOK {
		passed = false
	}

	if challenge.ImageID != nil {
		imageRef, err := s.resolveAvailableImageRef(ctx, *challenge.ImageID)
		if err != nil {
			passed = false
			*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
				Name:    "challenge_image",
				Passed:  false,
				Message: fmt.Sprintf("题目默认镜像不可用: %v", err),
			})
		} else {
			input.defaultImageRef = imageRef
			*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
				Name:    "challenge_image",
				Passed:  true,
				Message: "题目默认镜像可用",
			})
		}
	} else {
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "challenge_image",
			Passed:  true,
			Message: "题目未配置默认镜像",
		})
	}

	if s.topologyRepo == nil {
		return input, false, apperror.ErrInternal.WithCause(errors.New("challenge topology repository is not configured"))
	}
	topology, err := s.topologyRepo.FindChallengeTopologyByChallengeID(ctx, challenge.ID)
	if err != nil {
		if !errors.Is(err, challengeports.ErrChallengeTopologyNotFound) {
			return input, false, err
		}
		if challenge.ImageID == nil {
			if strings.TrimSpace(challenge.AttachmentURL) != "" {
				input.skipRuntime = true
				*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
					Name:    "topology_or_single_container",
					Passed:  true,
					Message: "题目仅提供附件内容，无需执行真实拉起",
				})
			} else {
				passed = false
				*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
					Name:    "topology_or_single_container",
					Passed:  false,
					Message: "未配置拓扑且题目默认镜像为空，无法执行真实拉起",
				})
			}
		} else {
			*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
				Name:    "topology_or_single_container",
				Passed:  true,
				Message: "未配置拓扑，将按单容器模式自测",
			})
		}
		return input, passed, nil
	}

	spec, err := challengecontracts.DecodeTopologySpec(topology.Spec)
	if err != nil {
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_spec",
			Passed:  false,
			Message: fmt.Sprintf("拓扑解码失败: %v", err),
		})
		return input, false, nil
	}
	if len(spec.Nodes) == 0 {
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_spec",
			Passed:  false,
			Message: "拓扑至少需要一个节点",
		})
		return input, false, nil
	}

	entryKey := strings.TrimSpace(topology.EntryNodeKey)
	if entryKey == "" {
		entryKey = spec.Nodes[0].Key
	}
	entryPortOK := false
	needsDefaultImage := false
	for _, node := range spec.Nodes {
		if node.Key == entryKey && node.ServicePort > 0 {
			entryPortOK = true
		}
		if node.ImageID == 0 {
			needsDefaultImage = true
			continue
		}
		if _, exists := input.nodeImageRefs[node.ImageID]; exists {
			continue
		}
		nodeImageRef, resolveErr := s.resolveAvailableImageRef(ctx, node.ImageID)
		if resolveErr != nil {
			passed = false
			*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
				Name:    "topology_images",
				Passed:  false,
				Message: fmt.Sprintf("拓扑节点镜像不可用 (image_id=%d): %v", node.ImageID, resolveErr),
			})
			break
		}
		input.nodeImageRefs[node.ImageID] = nodeImageRef
	}
	if passed {
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_images",
			Passed:  true,
			Message: "拓扑节点镜像检查通过",
		})
	}

	if needsDefaultImage && input.defaultImageRef == "" {
		passed = false
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_default_image",
			Passed:  false,
			Message: "拓扑存在未指定 image_id 的节点，但题目默认镜像不可用",
		})
	} else {
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_default_image",
			Passed:  true,
			Message: "拓扑默认镜像策略检查通过",
		})
	}

	if !entryPortOK {
		passed = false
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_entry",
			Passed:  false,
			Message: "拓扑入口节点不存在或未设置 service_port",
		})
	} else {
		*steps = append(*steps, challengecontracts.ChallengeSelfCheckStepResp{
			Name:    "topology_entry",
			Passed:  true,
			Message: "拓扑入口节点配置有效",
		})
	}

	input.useTopology = true
	input.topologySpec = spec
	input.entryNodeKey = entryKey
	return input, passed, nil
}

func (s *ChallengeSelfCheckService) validateFlagConfig(challenge *challengeports.ChallengeWriteModel) (bool, string) {
	switch challenge.FlagType {
	case challengecontracts.FlagTypeStatic:
		if challenge.FlagHash == "" || challenge.FlagSalt == "" {
			return false, "静态 Flag 未正确配置（缺少 hash/salt）"
		}
		return true, "静态 Flag 配置有效"
	case challengecontracts.FlagTypeDynamic:
		if strings.TrimSpace(s.selfCheckCfg.FlagGlobalSecret) == "" {
			return false, "动态 Flag 依赖的全局密钥未配置"
		}
		return true, "动态 Flag 配置有效"
	case challengecontracts.FlagTypeRegex:
		if _, err := regexp.Compile(strings.TrimSpace(challenge.FlagRegex)); err != nil {
			return false, fmt.Sprintf("Regex Flag 配置无效: %v", err)
		}
		return true, "Regex Flag 配置有效"
	case challengecontracts.FlagTypeManualReview:
		return true, "人工审核题已跳过 Flag 自动校验"
	default:
		return false, "Flag 类型无效"
	}
}

func (s *ChallengeSelfCheckService) buildRuntimeFlag(challenge *challengeports.ChallengeWriteModel) (string, error) {
	switch challenge.FlagType {
	case challengecontracts.FlagTypeStatic:
		return challenge.FlagHash, nil
	case challengecontracts.FlagTypeDynamic:
		nonce, err := randomstring.Generate()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(s.selfCheckCfg.FlagGlobalSecret) == "" {
			return "", fmt.Errorf("flag global secret is empty")
		}
		return crypto.GenerateDynamicFlag(0, challenge.ID, s.selfCheckCfg.FlagGlobalSecret, nonce, challenge.FlagPrefix), nil
	case challengecontracts.FlagTypeRegex, challengecontracts.FlagTypeManualReview:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported flag type %s", challenge.FlagType)
	}
}

func (s *ChallengeSelfCheckService) buildTopologyRuntimeRequest(
	input challengeSelfCheckRuntimeInput,
	flag string,
) (*challengeports.RuntimeTopologyCreateRequest, error) {
	req := &challengeports.RuntimeTopologyCreateRequest{
		Networks: make([]challengeports.RuntimeTopologyCreateNetwork, 0, len(input.topologySpec.Networks)),
		Nodes:    make([]challengeports.RuntimeTopologyCreateNode, 0, len(input.topologySpec.Nodes)),
		Policies: append([]challengecontracts.TopologyTrafficPolicy(nil), input.topologySpec.Policies...),
	}
	for _, network := range input.topologySpec.Networks {
		req.Networks = append(req.Networks, challengeports.RuntimeTopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	defaultNetworkKey := challengecontracts.TopologyDefaultNetworkKey
	if len(req.Networks) > 0 {
		defaultNetworkKey = req.Networks[0].Key
	}

	for _, node := range input.topologySpec.Nodes {
		imageRef := input.defaultImageRef
		if node.ImageID > 0 {
			imageRef = input.nodeImageRefs[node.ImageID]
		}
		if imageRef == "" {
			return nil, fmt.Errorf("node %s image is empty", node.Key)
		}

		env := make(map[string]string, len(node.Env)+1)
		for key, value := range node.Env {
			env[key] = value
		}
		if node.InjectFlag {
			env["FLAG"] = flag
		}

		networkKeys := append([]string(nil), node.NetworkKeys...)
		if len(networkKeys) == 0 {
			networkKeys = []string{defaultNetworkKey}
		}

		var resources *runtimecontracts.ResourceLimits
		if node.Resources != nil {
			resources = &runtimecontracts.ResourceLimits{
				CPUQuota:  node.Resources.CPUQuota,
				Memory:    node.Resources.MemoryMB * 1024 * 1024,
				PidsLimit: node.Resources.PidsLimit,
			}
		}

		req.Nodes = append(req.Nodes, challengeports.RuntimeTopologyCreateNode{
			Key:             node.Key,
			Image:           imageRef,
			Env:             env,
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.Key == input.entryNodeKey,
			NetworkKeys:     networkKeys,
			Resources:       resources,
		})
	}
	if len(req.Networks) == 0 {
		req.Networks = []challengeports.RuntimeTopologyCreateNetwork{
			{Key: challengecontracts.TopologyDefaultNetworkKey},
		}
	}
	return req, nil
}

func (s *ChallengeSelfCheckService) resolveAvailableImageRef(ctx context.Context, imageID int64) (string, error) {
	if imageID <= 0 {
		return "", fmt.Errorf("invalid image id")
	}
	imageItem, err := s.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeImageNotFound) {
			return "", apperror.ErrNotFound.WithCause(errors.New(domain.ErrMsgImageNotFound))
		}
		return "", err
	}
	if imageItem.Status != challengeentity.ImageStatusAvailable {
		return "", fmt.Errorf("image %d status=%s", imageItem.ID, imageItem.Status)
	}
	return challengeentity.BuildRuntimeImageRef(imageItem), nil
}
