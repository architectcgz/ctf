package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type writeupServiceRawRepository interface {
	challengeports.ChallengeWriteupChallengeLookupRepository
	challengeports.ChallengeWriteupUserLookupRepository
	challengeports.ChallengeAdminWriteupRepository
	challengeports.ChallengeReleasedWriteupRepository
	challengeports.ChallengeWriteupSolveStatusRepository
	challengeports.ChallengeSubmissionWriteupRepository
	challengeports.ChallengeTeacherSubmissionWriteupRepository
	challengeports.ChallengeSolutionQueryRepository
}

type WriteupServiceRepository struct {
	raw writeupServiceRawRepository
}

func NewWriteupServiceRepository(raw writeupServiceRawRepository) *WriteupServiceRepository {
	return &WriteupServiceRepository{raw: raw}
}

func (r *WriteupServiceRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	item, err := r.raw.FindByID(ctx, id)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeWriteupChallengeNotFound)
}

func (r *WriteupServiceRepository) FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
	item, err := r.raw.FindUserByID(ctx, userID)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeWriteupRequesterNotFound)
}

func (r *WriteupServiceRepository) FindWriteupByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengeWriteup, error) {
	item, err := r.raw.FindWriteupByChallengeID(ctx, challengeID)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeOfficialWriteupNotFound)
}

func (r *WriteupServiceRepository) UpsertWriteup(ctx context.Context, writeup *challengeentity.ChallengeWriteup) error {
	return r.raw.UpsertWriteup(ctx, writeup)
}

func (r *WriteupServiceRepository) DeleteWriteupByChallengeID(ctx context.Context, challengeID int64) error {
	return r.raw.DeleteWriteupByChallengeID(ctx, challengeID)
}

func (r *WriteupServiceRepository) FindReleasedWriteupByChallengeID(ctx context.Context, challengeID int64, now time.Time) (*challengeentity.ChallengeWriteup, error) {
	item, err := r.raw.FindReleasedWriteupByChallengeID(ctx, challengeID, now)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeReleasedWriteupNotFound)
}

func (r *WriteupServiceRepository) GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error) {
	return r.raw.GetSolvedStatus(ctx, userID, challengeID)
}

func (r *WriteupServiceRepository) FindSubmissionWriteupByUserChallenge(ctx context.Context, userID, challengeID int64) (*challengeentity.SubmissionWriteup, error) {
	item, err := r.raw.FindSubmissionWriteupByUserChallenge(ctx, userID, challengeID)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeSubmissionWriteupNotFound)
}

func (r *WriteupServiceRepository) FindSubmissionWriteupByID(ctx context.Context, id int64) (*challengeentity.SubmissionWriteup, error) {
	item, err := r.raw.FindSubmissionWriteupByID(ctx, id)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeSubmissionWriteupDetailNotFound)
}

func (r *WriteupServiceRepository) UpsertSubmissionWriteup(ctx context.Context, writeup *challengeentity.SubmissionWriteup) error {
	return r.raw.UpsertSubmissionWriteup(ctx, writeup)
}

func (r *WriteupServiceRepository) GetTeacherSubmissionWriteupByID(ctx context.Context, id int64) (*challengeports.TeacherSubmissionWriteupRecord, error) {
	item, err := r.raw.GetTeacherSubmissionWriteupByID(ctx, id)
	return item, mapWriteupNotFound(err, challengeports.ErrChallengeTeacherSubmissionWriteupNotFound)
}

func (r *WriteupServiceRepository) ListTeacherSubmissionWriteups(ctx context.Context, query *challengecontracts.TeacherSubmissionWriteupQuery) ([]challengeports.TeacherSubmissionWriteupRecord, int64, error) {
	return r.raw.ListTeacherSubmissionWriteups(ctx, query)
}

func (r *WriteupServiceRepository) ListRecommendedSolutionsByChallengeID(ctx context.Context, challengeID int64, now time.Time) ([]challengeports.RecommendedSolutionRecord, error) {
	return r.raw.ListRecommendedSolutionsByChallengeID(ctx, challengeID, now)
}

func (r *WriteupServiceRepository) ListCommunitySolutionsByChallengeID(ctx context.Context, challengeID int64, query *challengecontracts.CommunityChallengeSolutionQuery) ([]challengeports.CommunitySolutionRecord, int64, error) {
	return r.raw.ListCommunitySolutionsByChallengeID(ctx, challengeID, query)
}

func mapWriteupNotFound(err error, sentinel error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sentinel
	}
	return err
}
