package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type stubPracticeRepository struct {
	withinInstanceStartTxFn                func(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceStartTxRepository) error) error
	withinInstanceRestartTxFn              func(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceRestartTxRepository) error) error
	withinAWDServiceOperationTxFn          func(ctx context.Context, fn func(txRepo practiceports.PracticeAWDServiceOperationTxRepository) error) error
	findContestByIDFn                      func(ctx context.Context, contestID int64) (*model.Contest, error)
	listDesiredRuntimeAWDContestsFn        func(ctx context.Context) ([]*model.Contest, error)
	findContestChallengeFn                 func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	findContestAWDServiceFn                func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	findContestAWDServiceRuntimeSubjectFn  func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error)
	listContestAWDServicesFn               func(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error)
	listContestAWDInstancesFn              func(ctx context.Context, contestID int64) ([]*model.Instance, error)
	findContestTeamFn                      func(ctx context.Context, contestID, teamID int64) (*model.Team, error)
	listContestTeamsFn                     func(ctx context.Context, contestID int64) ([]*model.Team, error)
	findContestRegistrationFn              func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error)
	listContestAWDScopeControlsFn          func(ctx context.Context, contestID int64) ([]*model.AWDScopeControl, error)
	listScopeAWDScopeControlsFn            func(ctx context.Context, contestID, teamID, serviceID int64) ([]*model.AWDScopeControl, error)
	upsertAWDScopeControlFn                func(ctx context.Context, control *model.AWDScopeControl) error
	deleteAWDScopeControlFn                func(ctx context.Context, contestID, teamID int64, scopeType, controlType string, serviceID int64) error
	lockInstanceScopeFn                    func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error
	findScopedExistingInstanceFn           func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	findScopedRestartableInstanceFn        func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error)
	countScopedRunningInstancesFn          func(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error)
	refreshInstanceExpiryFn                func(instanceID int64, expiresAt time.Time) error
	refreshInstanceExpiryWithContextFn     func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	resetInstanceRuntimeForRestartFn       func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error
	isHostPortReusableForRestartFn         func(ctx context.Context, instanceID int64, hostPort int) (bool, error)
	createInstanceFn                       func(ctx context.Context, instance *model.Instance) error
	createAWDServiceOperationFn            func(ctx context.Context, operation *model.AWDServiceOperation) error
	finishAWDServiceOperationFn            func(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error
	reserveAvailablePortFn                 func(ctx context.Context, start, end int) (int, error)
	reserveAvailablePortExcludingFn        func(ctx context.Context, start, end, excludedPort int) (int, error)
	bindReservedPortFn                     func(ctx context.Context, port int, instanceID int64) error
	releaseReservedPortFn                  func(ctx context.Context, port int) error
	releasePortForInstanceFn               func(ctx context.Context, port int, instanceID int64) error
	createSubmissionFn                     func(ctx context.Context, submission *model.Submission) error
	findCorrectSubmissionFn                func(ctx context.Context, userID, challengeID int64) (*model.Submission, error)
	listChallengeSubmissionsFn             func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error)
	updateSubmissionFn                     func(ctx context.Context, submission *model.Submission) error
	findUserByIDFn                         func(ctx context.Context, userID int64) (*model.User, error)
	listTeacherManualReviewSubmissionsFn   func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error)
	getTeacherManualReviewSubmissionByIDFn func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error)
	isUniqueViolationFn                    func(err error) bool
}

func (s *stubPracticeRepository) WithinInstanceStartTx(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceStartTxRepository) error) error {
	if s.withinInstanceStartTxFn != nil {
		return s.withinInstanceStartTxFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) WithinInstanceRestartTx(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceRestartTxRepository) error) error {
	if s.withinInstanceRestartTxFn != nil {
		return s.withinInstanceRestartTxFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) WithinAWDServiceOperationTx(ctx context.Context, fn func(txRepo practiceports.PracticeAWDServiceOperationTxRepository) error) error {
	if s.withinAWDServiceOperationTxFn != nil {
		return s.withinAWDServiceOperationTxFn(ctx, fn)
	}
	return fn(s)
}

func (s *stubPracticeRepository) FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error) {
	if s.findContestByIDFn != nil {
		return s.findContestByIDFn(ctx, contestID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListDesiredRuntimeAWDContests(ctx context.Context) ([]*model.Contest, error) {
	if s.listDesiredRuntimeAWDContestsFn != nil {
		return s.listDesiredRuntimeAWDContestsFn(ctx)
	}
	return []*model.Contest{}, nil
}

func (s *stubPracticeRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	if s.findContestChallengeFn != nil {
		return s.findContestChallengeFn(ctx, contestID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	if s.findContestAWDServiceFn != nil {
		return s.findContestAWDServiceFn(ctx, contestID, serviceID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) FindContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
	if s.findContestAWDServiceRuntimeSubjectFn != nil {
		return s.findContestAWDServiceRuntimeSubjectFn(ctx, contestID, serviceID)
	}
	service, err := s.FindContestAWDService(ctx, contestID, serviceID)
	if err != nil || service == nil {
		return nil, err
	}
	return stubContestAWDServiceRuntimeSubject(service)
}

func (s *stubPracticeRepository) ListContestAWDServices(ctx context.Context, contestID int64) ([]*model.ContestAWDService, error) {
	if s.listContestAWDServicesFn != nil {
		return s.listContestAWDServicesFn(ctx, contestID)
	}
	return []*model.ContestAWDService{}, nil
}

func (s *stubPracticeRepository) ListContestAWDInstances(ctx context.Context, contestID int64) ([]*model.Instance, error) {
	if s.listContestAWDInstancesFn != nil {
		return s.listContestAWDInstancesFn(ctx, contestID)
	}
	return []*model.Instance{}, nil
}

func (s *stubPracticeRepository) FindContestTeam(ctx context.Context, contestID, teamID int64) (*model.Team, error) {
	if s.findContestTeamFn != nil {
		return s.findContestTeamFn(ctx, contestID, teamID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListContestTeams(ctx context.Context, contestID int64) ([]*model.Team, error) {
	if s.listContestTeamsFn != nil {
		return s.listContestTeamsFn(ctx, contestID)
	}
	return []*model.Team{}, nil
}

func (s *stubPracticeRepository) FindContestRegistration(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
	if s.findContestRegistrationFn != nil {
		return s.findContestRegistrationFn(ctx, contestID, userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListContestAWDScopeControls(ctx context.Context, contestID int64) ([]*model.AWDScopeControl, error) {
	if s.listContestAWDScopeControlsFn != nil {
		return s.listContestAWDScopeControlsFn(ctx, contestID)
	}
	return nil, nil
}

func (s *stubPracticeRepository) ListScopeAWDScopeControls(ctx context.Context, contestID, teamID, serviceID int64) ([]*model.AWDScopeControl, error) {
	if s.listScopeAWDScopeControlsFn != nil {
		return s.listScopeAWDScopeControlsFn(ctx, contestID, teamID, serviceID)
	}
	return nil, nil
}

func (s *stubPracticeRepository) UpsertAWDScopeControl(ctx context.Context, control *model.AWDScopeControl) error {
	if s.upsertAWDScopeControlFn != nil {
		return s.upsertAWDScopeControlFn(ctx, control)
	}
	return nil
}

func (s *stubPracticeRepository) DeleteAWDScopeControl(ctx context.Context, contestID, teamID int64, scopeType, controlType string, serviceID int64) error {
	if s.deleteAWDScopeControlFn != nil {
		return s.deleteAWDScopeControlFn(ctx, contestID, teamID, scopeType, controlType, serviceID)
	}
	return nil
}

func (s *stubPracticeRepository) LockInstanceScope(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
	if s.lockInstanceScopeFn != nil {
		return s.lockInstanceScopeFn(ctx, userID, challengeID, scope)
	}
	return nil
}

func (s *stubPracticeRepository) FindScopedExistingInstance(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	if s.findScopedExistingInstanceFn != nil {
		return s.findScopedExistingInstanceFn(ctx, userID, challengeID, scope)
	}
	return nil, nil
}

func (s *stubPracticeRepository) FindScopedRestartableInstance(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	if s.findScopedRestartableInstanceFn != nil {
		return s.findScopedRestartableInstanceFn(ctx, userID, challengeID, scope)
	}
	return s.FindScopedExistingInstance(ctx, userID, challengeID, scope)
}

func (s *stubPracticeRepository) CountScopedRunningInstances(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error) {
	if s.countScopedRunningInstancesFn != nil {
		return s.countScopedRunningInstancesFn(ctx, userID, scope)
	}
	return 0, nil
}

func (s *stubPracticeRepository) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	if s.refreshInstanceExpiryWithContextFn != nil {
		return s.refreshInstanceExpiryWithContextFn(ctx, instanceID, expiresAt)
	}
	return nil
}

func (s *stubPracticeRepository) ResetInstanceRuntimeForRestart(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
	if s.resetInstanceRuntimeForRestartFn != nil {
		return s.resetInstanceRuntimeForRestartFn(ctx, instanceID, status, expiresAt, preserveHostPort)
	}
	return nil
}

func (s *stubPracticeRepository) IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
	if s.isHostPortReusableForRestartFn != nil {
		return s.isHostPortReusableForRestartFn(ctx, instanceID, hostPort)
	}
	return true, nil
}

func (s *stubPracticeRepository) CreateInstance(ctx context.Context, instance *model.Instance) error {
	if s.createInstanceFn != nil {
		return s.createInstanceFn(ctx, instance)
	}
	return nil
}

func stubContestAWDServiceRuntimeSubject(service *model.ContestAWDService) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
	if service == nil {
		return nil, nil
	}

	subject := &practiceports.ContestAWDServiceRuntimeSubject{
		ServiceID:     service.ID,
		ChallengeID:   service.AWDChallengeID,
		Visible:       service.IsVisible,
		SeedSignature: buildStubContestAWDServiceSeedSignature(service.ServiceSnapshot),
		RuntimeChallenge: &model.Challenge{
			ID:     service.AWDChallengeID,
			Status: model.ChallengeStatusPublished,
		},
	}

	snapshot, err := model.DecodeContestAWDServiceSnapshot(service.ServiceSnapshot)
	if err != nil {
		return nil, err
	}
	subject.RuntimeChallenge.Title = firstNonEmptyStubRuntimeValue(service.DisplayName, snapshot.Name)
	subject.RuntimeChallenge.Category = snapshot.Category
	subject.RuntimeChallenge.Difficulty = snapshot.Difficulty
	subject.RuntimeChallenge.Points = stubContestAWDServiceSnapshotPoints(service.ScoreConfig)
	subject.RuntimeChallenge.ImageID = stubContestAWDServiceSnapshotImageID(snapshot.RuntimeConfig)
	subject.RuntimeChallenge.FlagType = stubContestAWDServiceSnapshotFlagType(snapshot.FlagConfig)
	subject.RuntimeChallenge.FlagPrefix = stubContestAWDServiceSnapshotFlagPrefix(snapshot.FlagConfig)
	subject.RuntimeChallenge.InstanceSharing = stubContestAWDServiceSnapshotInstanceSharing(snapshot.RuntimeConfig)
	if subject.RuntimeChallenge.FlagPrefix == "" {
		subject.RuntimeChallenge.FlagPrefix = "flag"
	}
	subject.WorkspaceConfig = stubContestAWDDefenseWorkspaceConfig(snapshot.RuntimeConfig)

	topologyPayload, ok := snapshot.RuntimeConfig["topology"]
	if !ok {
		return subject, nil
	}
	topologyMap, ok := topologyPayload.(map[string]any)
	if !ok {
		return subject, nil
	}
	specPayload, ok := topologyMap["spec"]
	if !ok {
		return subject, nil
	}
	specRaw, err := json.Marshal(specPayload)
	if err != nil {
		return nil, err
	}
	entryNodeKey, _ := topologyMap["entry_node_key"].(string)
	subject.RuntimeTopology = &model.ChallengeTopology{
		ChallengeID:  service.AWDChallengeID,
		EntryNodeKey: entryNodeKey,
		Spec:         string(specRaw),
	}
	return subject, nil
}

func firstNonEmptyStubRuntimeValue(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func stubContestAWDServiceSnapshotPoints(scoreConfig string) int {
	if scoreConfig == "" {
		return 0
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(scoreConfig), &payload); err != nil {
		return 0
	}
	return stubContestAWDServiceSnapshotInt(payload["points"])
}

func stubContestAWDServiceSnapshotImageID(runtimeConfig map[string]any) int64 {
	if runtimeConfig == nil {
		return 0
	}
	value := stubContestAWDServiceSnapshotInt(runtimeConfig["image_id"])
	if value <= 0 {
		return 0
	}
	return int64(value)
}

func stubContestAWDServiceSnapshotFlagType(flagConfig map[string]any) string {
	if flagConfig == nil {
		return model.FlagTypeDynamic
	}
	value, _ := flagConfig["flag_type"].(string)
	if value == "" {
		return model.FlagTypeDynamic
	}
	return value
}

func stubContestAWDServiceSnapshotFlagPrefix(flagConfig map[string]any) string {
	if flagConfig == nil {
		return "flag"
	}
	value, _ := flagConfig["flag_prefix"].(string)
	if value == "" {
		return "flag"
	}
	return value
}

func stubContestAWDServiceSnapshotInstanceSharing(runtimeConfig map[string]any) model.InstanceSharing {
	if runtimeConfig == nil {
		return model.InstanceSharingPerTeam
	}
	value, _ := runtimeConfig["instance_sharing"].(string)
	switch model.InstanceSharing(value) {
	case model.InstanceSharingShared:
		return model.InstanceSharingShared
	case model.InstanceSharingPerUser:
		return model.InstanceSharingPerUser
	case model.InstanceSharingPerTeam:
		return model.InstanceSharingPerTeam
	default:
		return model.InstanceSharingPerTeam
	}
}

func stubContestAWDServiceSnapshotInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		next, err := typed.Int64()
		if err != nil {
			return 0
		}
		return int(next)
	default:
		return 0
	}
}

func stubContestAWDDefenseWorkspaceConfig(runtimeConfig map[string]any) *practiceports.ContestAWDDefenseWorkspaceConfig {
	if runtimeConfig == nil {
		return nil
	}
	raw, ok := runtimeConfig["defense_workspace"]
	if !ok {
		return nil
	}
	payload, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	seedRoot := stubReadRuntimeString(payload["seed_root"])
	workspaceRoots := stubReadRuntimeStringList(payload["workspace_roots"])
	if seedRoot == "" || len(workspaceRoots) == 0 {
		return nil
	}
	writableRoots := make(map[string]struct{}, len(workspaceRoots))
	for _, root := range stubReadRuntimeStringList(payload["writable_roots"]) {
		writableRoots[root] = struct{}{}
	}
	roots := make([]practiceports.ContestAWDDefenseWorkspaceRoot, 0, len(workspaceRoots))
	for _, root := range workspaceRoots {
		_, writable := writableRoots[root]
		roots = append(roots, practiceports.ContestAWDDefenseWorkspaceRoot{
			Source:   root,
			ReadOnly: !writable,
		})
	}
	runtimeMounts := stubContestAWDDefenseRuntimeMounts(payload["runtime_mounts"])
	if len(runtimeMounts) == 0 {
		return nil
	}
	return &practiceports.ContestAWDDefenseWorkspaceConfig{
		SeedRoot:        seedRoot,
		WorkspaceRoots:  roots,
		RuntimeMounts:   runtimeMounts,
		CheckerTokenEnv: stubReadRuntimeString(runtimeConfig["checker_token_env"]),
	}
}

func stubContestAWDDefenseRuntimeMounts(raw any) []practiceports.ContestAWDDefenseRuntimeMount {
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	result := make([]practiceports.ContestAWDDefenseRuntimeMount, 0, len(items))
	for _, item := range items {
		payload, ok := item.(map[string]any)
		if !ok {
			return nil
		}
		source := stubReadRuntimeString(payload["source"])
		target := stubReadRuntimeString(payload["target"])
		mode := stubReadRuntimeString(payload["mode"])
		if source == "" || target == "" || mode == "" {
			return nil
		}
		result = append(result, practiceports.ContestAWDDefenseRuntimeMount{
			Source:   source,
			Target:   target,
			ReadOnly: mode == "ro",
		})
	}
	return result
}

func stubReadRuntimeString(raw any) string {
	value, _ := raw.(string)
	return value
}

func stubReadRuntimeStringList(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			value := stubReadRuntimeString(item)
			if value != "" {
				items = append(items, value)
			}
		}
		return items
	default:
		return nil
	}
}

func buildStubContestAWDServiceSeedSignature(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func (s *stubPracticeRepository) CreateAWDServiceOperation(ctx context.Context, operation *model.AWDServiceOperation) error {
	if s.createAWDServiceOperationFn != nil {
		return s.createAWDServiceOperationFn(ctx, operation)
	}
	return nil
}

func (s *stubPracticeRepository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if s.finishAWDServiceOperationFn != nil {
		return s.finishAWDServiceOperationFn(ctx, operationID, status, errorMessage, finishedAt)
	}
	return nil
}

func (s *stubPracticeRepository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	if s.reserveAvailablePortFn != nil {
		return s.reserveAvailablePortFn(ctx, start, end)
	}
	return start, nil
}

func (s *stubPracticeRepository) ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error) {
	if s.reserveAvailablePortExcludingFn != nil {
		return s.reserveAvailablePortExcludingFn(ctx, start, end, excludedPort)
	}
	return s.ReserveAvailablePort(ctx, start, end)
}

func (s *stubPracticeRepository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	if s.bindReservedPortFn != nil {
		return s.bindReservedPortFn(ctx, port, instanceID)
	}
	return nil
}

func (s *stubPracticeRepository) ReleaseReservedPort(ctx context.Context, port int) error {
	if s.releaseReservedPortFn != nil {
		return s.releaseReservedPortFn(ctx, port)
	}
	return nil
}

func (s *stubPracticeRepository) ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error {
	if s.releasePortForInstanceFn != nil {
		return s.releasePortForInstanceFn(ctx, port, instanceID)
	}
	return nil
}

func (s *stubPracticeRepository) CreateSubmission(ctx context.Context, submission *model.Submission) error {
	if s.createSubmissionFn != nil {
		return s.createSubmissionFn(ctx, submission)
	}
	return nil
}

func (s *stubPracticeRepository) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	if s.findCorrectSubmissionFn != nil {
		return s.findCorrectSubmissionFn(ctx, userID, challengeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListChallengeSubmissions(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
	if s.listChallengeSubmissionsFn != nil {
		return s.listChallengeSubmissionsFn(ctx, userID, challengeID, limit)
	}
	return nil, nil
}

func (s *stubPracticeRepository) UpdateSubmission(ctx context.Context, submission *model.Submission) error {
	if s.updateSubmissionFn != nil {
		return s.updateSubmissionFn(ctx, submission)
	}
	return nil
}

func (s *stubPracticeRepository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(ctx, userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) ListTeacherManualReviewSubmissions(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	if s.listTeacherManualReviewSubmissionsFn != nil {
		return s.listTeacherManualReviewSubmissionsFn(ctx, query)
	}
	return nil, 0, nil
}

func (s *stubPracticeRepository) GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	if s.getTeacherManualReviewSubmissionByIDFn != nil {
		return s.getTeacherManualReviewSubmissionByIDFn(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubPracticeRepository) IsUniqueViolation(err error) bool {
	if s.isUniqueViolationFn != nil {
		return s.isUniqueViolationFn(err)
	}
	return false
}
