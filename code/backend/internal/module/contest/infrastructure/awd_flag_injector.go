package infrastructure

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type noopAWDFlagInjector struct {
	log *zap.Logger
}

func NewNoopAWDFlagInjector(log *zap.Logger) contestports.AWDFlagInjector {
	if log == nil {
		log = zap.NewNop()
	}
	return &noopAWDFlagInjector{log: log}
}

func (i *noopAWDFlagInjector) InjectRoundFlags(_ context.Context, contest *model.Contest, round *model.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	i.log.Debug("skip_awd_flag_injection",
		zap.Int64("contest_id", contest.ID),
		zap.Int64("round_id", round.ID),
		zap.Int("assignment_count", len(assignments)),
	)
	return nil
}

type dockerAWDFlagInjector struct {
	db           *gorm.DB
	writer       contestports.AWDContainerFileWriter
	flagFilePath string
	log          *zap.Logger
}

func NewDockerAWDFlagInjector(db *gorm.DB, writer contestports.AWDContainerFileWriter, log *zap.Logger) contestports.AWDFlagInjector {
	if writer == nil {
		return NewNoopAWDFlagInjector(log)
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &dockerAWDFlagInjector{
		db:           db,
		writer:       writer,
		flagFilePath: "/flag/flag.txt",
		log:          log,
	}
}

func (i *dockerAWDFlagInjector) InjectRoundFlags(ctx context.Context, contest *model.Contest, round *model.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	if i.db == nil || contest == nil || round == nil || len(assignments) == 0 {
		return nil
	}

	type pair struct {
		teamID      int64
		challengeID int64
	}
	seen := make(map[pair]struct{}, len(assignments))
	for _, item := range assignments {
		key := pair{teamID: item.TeamID, challengeID: item.ChallengeID}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		containerIDs, err := i.findTargetContainers(ctx, contest.ID, item.TeamID, item.ChallengeID)
		if err != nil {
			return err
		}
		for _, containerID := range containerIDs {
			if err := i.writer.WriteFileToContainer(ctx, containerID, i.flagFilePath, []byte(item.Flag)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *dockerAWDFlagInjector) findTargetContainers(ctx context.Context, contestID, teamID, challengeID int64) ([]string, error) {
	var instances []model.Instance
	if err := i.db.WithContext(ctx).
		Table("instances AS inst").
		Select("inst.*").
		Where("inst.challenge_id = ?", challengeID).
		Where("inst.status = ?", model.InstanceStatusRunning).
		Where(
			"(inst.contest_id = ? AND inst.team_id = ?) OR ("+
				"inst.team_id IS NULL AND EXISTS ("+
				"SELECT 1 FROM team_members AS tm "+
				"WHERE tm.contest_id = ? AND tm.team_id = ? AND tm.user_id = inst.user_id))",
			contestID, teamID, contestID, teamID,
		).
		Order("inst.id ASC").
		Scan(&instances).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(instances))
	containerIDs := make([]string, 0, len(instances))
	for _, instance := range instances {
		for _, containerID := range collectInstanceContainerIDs(&instance) {
			if _, exists := seen[containerID]; exists || containerID == "" {
				continue
			}
			seen[containerID] = struct{}{}
			containerIDs = append(containerIDs, containerID)
		}
	}
	return containerIDs, nil
}

func collectInstanceContainerIDs(instance *model.Instance) []string {
	if instance == nil {
		return nil
	}
	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil || len(details.Containers) == 0 {
		if instance.ContainerID == "" {
			return nil
		}
		return []string{instance.ContainerID}
	}

	ids := make([]string, 0, len(details.Containers))
	for _, item := range details.Containers {
		if item.ContainerID != "" {
			ids = append(ids, item.ContainerID)
		}
	}
	if len(ids) == 0 && instance.ContainerID != "" {
		return []string{instance.ContainerID}
	}
	return ids
}
