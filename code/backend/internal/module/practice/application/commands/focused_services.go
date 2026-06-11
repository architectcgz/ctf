package commands

import (
	"context"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
	"go.uber.org/zap"
)

type CommandServices struct {
	core          *serviceCore
	Instances     *InstanceLifecycleService
	Submissions   *SubmissionService
	ManualReviews *ManualReviewService
	Runtime       *RuntimeLifecycleService
}

func NewCommandServices(
	repo practiceCommandRepository,
	imageRepo challengecontracts.ImageStore,
	instanceRepo instanceRepository,
	runtimeService practiceports.RuntimeInstanceService,
	scoreService ScoreUpdater,
	rateLimitStore practiceports.PracticeFlagSubmitRateLimitStore,
	cfg *config.Config,
	logger *zap.Logger,
) *CommandServices {
	core := newServiceCore(repo, imageRepo, instanceRepo, runtimeService, scoreService, rateLimitStore, cfg, logger)
	return &CommandServices{
		core:          core,
		Instances:     NewInstanceLifecycleService(core),
		Submissions:   NewSubmissionService(core),
		ManualReviews: NewManualReviewService(core),
		Runtime:       NewRuntimeLifecycleService(core),
	}
}

func (s *CommandServices) SetEventBus(bus platformevents.Bus) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetEventBus(bus)
	}
	return s
}

func (s *CommandServices) SetDesiredAWDReconcileStateStore(store practiceports.PracticeDesiredAWDReconcileStateStore) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetDesiredAWDReconcileStateStore(store)
	}
	return s
}

func (s *CommandServices) SetSchedulerLockStore(store practiceports.PracticeInstanceSchedulerLockStore) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetSchedulerLockStore(store)
	}
	return s
}

func (s *CommandServices) SetInstanceReadinessProbe(probe practiceports.PracticeInstanceReadinessProbe) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetInstanceReadinessProbe(probe)
	}
	return s
}

func (s *CommandServices) SetContestScopeRepository(repo practiceports.PracticeContestScopeRepository) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetContestScopeRepository(repo)
	}
	return s
}

func (s *CommandServices) SetRuntimeSubjectRepository(repo practiceports.PracticeRuntimeSubjectRepository) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetRuntimeSubjectRepository(repo)
	}
	return s
}

func (s *CommandServices) SetRuntimeNodeSelector(selector practiceports.RuntimeNodeSelector) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetRuntimeNodeSelector(selector)
	}
	return s
}

func (s *CommandServices) SetManualReviewRepository(repo practiceports.PracticeManualReviewRepository) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetManualReviewRepository(repo)
	}
	return s
}

func (s *CommandServices) SetSolvedSubmissionRepository(repo practiceports.PracticeSolvedSubmissionRepository) *CommandServices {
	if s != nil && s.core != nil {
		s.core.SetSolvedSubmissionRepository(repo)
	}
	return s
}

type RuntimeLifecycleService struct {
	service *serviceCore
}

func NewRuntimeLifecycleService(service *serviceCore) *RuntimeLifecycleService {
	return &RuntimeLifecycleService{service: service}
}

func (s *RuntimeLifecycleService) StartBackgroundTasks(ctx context.Context) {
	if s == nil || s.service == nil {
		return
	}
	s.service.StartBackgroundTasks(ctx)
}

func (s *RuntimeLifecycleService) Close(ctx context.Context) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.Close(ctx)
}

func (s *RuntimeLifecycleService) RunProvisioningLoop(ctx context.Context) {
	if s == nil || s.service == nil {
		return
	}
	s.service.RunProvisioningLoop(ctx)
}

func (s *RuntimeLifecycleService) ReconcileDesiredAWDInstances(ctx context.Context) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.ReconcileDesiredAWDInstances(ctx)
}

type InstanceLifecycleService struct {
	service *serviceCore
}

func NewInstanceLifecycleService(service *serviceCore) *InstanceLifecycleService {
	return &InstanceLifecycleService{service: service}
}

func (s *InstanceLifecycleService) StartChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.InstanceResp, error) {
	return s.service.StartChallenge(ctx, userID, challengeID)
}

func (s *InstanceLifecycleService) StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*instancecontracts.InstanceResp, error) {
	return s.service.StartContestChallenge(ctx, userID, contestID, challengeID)
}

func (s *InstanceLifecycleService) StartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*instancecontracts.InstanceResp, error) {
	return s.service.StartContestAWDService(ctx, userID, contestID, serviceID)
}

func (s *InstanceLifecycleService) RestartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*instancecontracts.InstanceResp, error) {
	return s.service.RestartContestAWDService(ctx, userID, contestID, serviceID)
}

func (s *InstanceLifecycleService) GetContestAWDInstanceOrchestration(ctx context.Context, contestID int64) (*practicecontracts.AdminAWDInstanceOrchestrationResp, error) {
	return s.service.GetContestAWDInstanceOrchestration(ctx, contestID)
}

func (s *InstanceLifecycleService) StartAdminContestAWDTeamService(ctx context.Context, contestID, teamID, serviceID int64) (*practicecontracts.AdminAWDInstanceItemResp, error) {
	return s.service.StartAdminContestAWDTeamService(ctx, contestID, teamID, serviceID)
}

func (s *InstanceLifecycleService) SetAdminContestAWDTeamRetired(ctx context.Context, contestID, teamID, actorUserID int64, retired bool, reason string) (*practicecontracts.AdminAWDScopeControlResp, error) {
	return s.service.SetAdminContestAWDTeamRetired(ctx, contestID, teamID, actorUserID, retired, reason)
}

func (s *InstanceLifecycleService) SetAdminContestAWDTeamServiceDisabled(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, disabled bool, reason string) (*practicecontracts.AdminAWDScopeControlResp, error) {
	return s.service.SetAdminContestAWDTeamServiceDisabled(ctx, contestID, teamID, serviceID, actorUserID, disabled, reason)
}

func (s *InstanceLifecycleService) SetAdminContestAWDDesiredReconcileSuppressed(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, suppressed bool, reason string) (*practicecontracts.AdminAWDScopeControlResp, error) {
	return s.service.SetAdminContestAWDDesiredReconcileSuppressed(ctx, contestID, teamID, serviceID, actorUserID, suppressed, reason)
}

func (s *InstanceLifecycleService) PrewarmAdminContestAWDInstances(ctx context.Context, contestID int64, teamID *int64) (*practicecontracts.AdminAWDInstancePrewarmResp, error) {
	return s.service.PrewarmAdminContestAWDInstances(ctx, contestID, teamID)
}

type SubmissionService struct {
	service *serviceCore
}

func NewSubmissionService(service *serviceCore) *SubmissionService {
	return &SubmissionService{service: service}
}

func (s *SubmissionService) SubmitFlag(ctx context.Context, userID, challengeID int64, flag string) (*practicecontracts.SubmissionResp, error) {
	return s.service.SubmitFlag(ctx, userID, challengeID, flag)
}

func (s *SubmissionService) ListMyChallengeSubmissions(ctx context.Context, userID, challengeID int64) ([]*practicecontracts.ChallengeSubmissionRecordResp, error) {
	return s.service.ListMyChallengeSubmissions(ctx, userID, challengeID)
}

type ManualReviewService struct {
	service *serviceCore
}

func NewManualReviewService(service *serviceCore) *ManualReviewService {
	return &ManualReviewService{service: service}
}

func (s *ManualReviewService) ListTeacherManualReviewSubmissions(ctx context.Context, requesterID int64, requesterRole string, query *practicecontracts.TeacherManualReviewSubmissionQuery) (*practicecontracts.PageResult[*practicecontracts.TeacherManualReviewSubmissionItemResp], error) {
	return s.service.ListTeacherManualReviewSubmissions(ctx, requesterID, requesterRole, query)
}

func (s *ManualReviewService) GetTeacherManualReviewSubmission(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*practicecontracts.TeacherManualReviewSubmissionDetailResp, error) {
	return s.service.GetTeacherManualReviewSubmission(ctx, submissionID, requesterID, requesterRole)
}

func (s *ManualReviewService) ReviewManualReviewSubmission(ctx context.Context, submissionID, reviewerID int64, reviewerRole string, req *practicecontracts.ReviewManualReviewSubmissionReq) (*practicecontracts.TeacherManualReviewSubmissionDetailResp, error) {
	return s.service.ReviewManualReviewSubmission(ctx, submissionID, reviewerID, reviewerRole, req)
}
