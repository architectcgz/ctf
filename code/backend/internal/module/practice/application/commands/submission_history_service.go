package commands

import (
	"context"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func (s *serviceCore) ListMyChallengeSubmissions(ctx context.Context, userID, challengeID int64) ([]*ChallengeSubmissionRecordResp, error) {
	if s.runtimeSubject == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice runtime subject repository is nil"))
	}
	challengeItem, err := s.runtimeSubject.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if challengeItem.Status != practiceentity.ChallengeStatusPublished {
		return nil, challengecontracts.ErrChallengeNotPublish
	}

	items, err := s.repo.ListChallengeSubmissions(ctx, userID, challengeID, 20)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	resp := make([]*ChallengeSubmissionRecordResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, challengeSubmissionRecordRespFromModel(item))
	}
	return resp, nil
}

func challengeSubmissionRecordRespFromModel(item practiceports.SubmissionRecord) *ChallengeSubmissionRecordResp {
	status := SubmissionStatusIncorrect
	answer := ""

	if item.ReviewStatus == practiceports.SubmissionReviewStatusPending {
		status = SubmissionStatusPendingReview
		answer = item.Flag
	} else if item.IsCorrect {
		status = SubmissionStatusCorrect
	}

	resp := practiceCommandResponseMapperInst.ToChallengeSubmissionRecordRespBasePtr(&challengeSubmissionRecordRespSource{
		ID:          item.ID,
		SubmittedAt: item.SubmittedAt,
	})
	resp.Status = status
	resp.Answer = answer
	return resp
}
