package composition

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type PracticeModule struct {
	BackgroundCloser asyncTaskCloser
	Handler          *practicehttp.Handler
}

type practiceModuleDeps struct {
	commandRepo    practiceports.PracticeCommandRepository
	scoreRepo      practiceports.PracticeScoreRepository
	rankingRepo    practiceports.PracticeRankingRepository
	instanceRepo   practiceports.InstanceRepository
	runtimeService practiceports.RuntimeInstanceService
	challengeRepo  practiceRuntimeChallengeContract
	imageStore     practiceRuntimeImageStore
	assessment     practiceRuntimeAssessmentService
}

type practiceRuntimeChallengeContract interface {
	FindByID(id int64) (*model.Challenge, error)
	FindHintByLevel(challengeID int64, level int) (*model.ChallengeHint, error)
	CreateHintUnlock(unlock *model.ChallengeHintUnlock) error
	FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error)
}

type practiceRuntimeImageStore interface {
	FindByID(id int64) (*model.Image, error)
}

type practiceRuntimeAssessmentService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

func BuildPracticeModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) *PracticeModule {
	cfg := root.Config()
	log := root.Logger()
	cache := root.Cache()
	deps := buildPracticeModuleDeps(root, challenge, runtime, assessment)
	scoreService := practicecmd.NewScoreService(deps.scoreRepo, cache, log.Named("score_service"), &cfg.Score)
	service := practicecmd.NewService(
		deps.commandRepo,
		deps.challengeRepo,
		deps.imageStore,
		deps.instanceRepo,
		deps.runtimeService,
		scoreService,
		deps.assessment,
		cache,
		cfg,
		log.Named("practice_service"),
	)
	service.SetEventBus(root.Events)

	return &PracticeModule{
		BackgroundCloser: service,
		Handler:          practicehttp.NewHandler(service),
	}
}

func buildPracticeModuleDeps(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) practiceModuleDeps {
	repo := practiceinfra.NewRepository(root.DB())
	return practiceModuleDeps{
		commandRepo:    repo,
		scoreRepo:      repo,
		rankingRepo:    repo,
		instanceRepo:   runtime.practice.instanceRepository,
		runtimeService: runtime.practice.runtimeService,
		challengeRepo:  challenge.Catalog,
		imageStore:     challenge.ImageStore,
		assessment:     assessment.ProfileService,
	}
}

type practiceRuntimeCleanerBridge interface {
	CleanupRuntime(instance *model.Instance) error
}

type practiceRuntimeRepositoryBridge interface {
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

type practiceRuntimeInstanceService interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type practiceRuntimeProvisioningBridge interface {
	CreateTopology(ctx context.Context, req *runtimeapp.TopologyCreateRequest) (*runtimeapp.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type practiceRuntimeInstanceServiceAdapter struct {
	cleaner     practiceRuntimeCleanerBridge
	provisioner practiceRuntimeProvisioningBridge
}

func newPracticeRuntimeInstanceServiceAdapter(cleaner practiceRuntimeCleanerBridge, provisioner practiceRuntimeProvisioningBridge) *practiceRuntimeInstanceServiceAdapter {
	if isNilPracticeRuntimeDependency(cleaner) && isNilPracticeRuntimeDependency(provisioner) {
		return nil
	}
	return &practiceRuntimeInstanceServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
	}
}

func isNilPracticeRuntimeDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (a *practiceRuntimeInstanceServiceAdapter) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(instance)
}

func (a *practiceRuntimeInstanceServiceAdapter) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if a == nil || a.provisioner == nil || req == nil {
		return nil, nil
	}

	result, err := a.provisioner.CreateTopology(ctx, toRuntimeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResult(result), nil
}

func (a *practiceRuntimeInstanceServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.provisioner == nil {
		return "", "", 0, 0, nil
	}
	return a.provisioner.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func toRuntimeTopologyCreateRequest(req *practiceports.TopologyCreateRequest) *runtimeapp.TopologyCreateRequest {
	if req == nil {
		return nil
	}

	networks := make([]runtimeapp.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeapp.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtimeapp.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeapp.TopologyCreateNode{
			Key:          node.Key,
			Image:        node.Image,
			Env:          cloneStringMap(node.Env),
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.IsEntryPoint,
			NetworkKeys:  append([]string(nil), node.NetworkKeys...),
			Resources:    cloneResourceLimits(node.Resources),
		})
	}

	return &runtimeapp.TopologyCreateRequest{
		Networks:         networks,
		Nodes:            nodes,
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	}
}

func fromRuntimeTopologyCreateResult(result *runtimeapp.TopologyCreateResult) *practiceports.TopologyCreateResult {
	if result == nil {
		return nil
	}
	return &practiceports.TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
