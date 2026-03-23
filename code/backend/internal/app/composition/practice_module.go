package composition

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	practiceModule "ctf-platform/internal/module/practice"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type PracticeModule struct {
	BackgroundCloser asyncTaskCloser
	Handler          *practiceModule.Handler
}

func BuildPracticeModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) *PracticeModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := practiceModule.NewRepository(db)
	scoreService := practiceModule.NewScoreService(repo, cache, log.Named("score_service"), &cfg.Score)
	service := practiceModule.NewService(
		repo,
		challenge.Catalog,
		challenge.ImageStore,
		runtime.practice.instanceRepository,
		runtime.practice.runtimeService,
		scoreService,
		assessment.ProfileService,
		cache,
		cfg,
		log.Named("practice_service"),
	)
	service.SetEventBus(root.Events)

	return &PracticeModule{
		BackgroundCloser: service,
		Handler:          practiceModule.NewHandler(service),
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
	CreateTopology(ctx context.Context, req *practiceModule.TopologyCreateRequest) (*practiceModule.TopologyCreateResult, error)
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

func (a *practiceRuntimeInstanceServiceAdapter) CreateTopology(ctx context.Context, req *practiceModule.TopologyCreateRequest) (*practiceModule.TopologyCreateResult, error) {
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

func toRuntimeTopologyCreateRequest(req *practiceModule.TopologyCreateRequest) *runtimeapp.TopologyCreateRequest {
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

func fromRuntimeTopologyCreateResult(result *runtimeapp.TopologyCreateResult) *practiceModule.TopologyCreateResult {
	if result == nil {
		return nil
	}
	return &practiceModule.TopologyCreateResult{
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
