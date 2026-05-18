package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/pkg/errcode"
)

func (s *Service) ListMyChallengeSubmissions(ctx context.Context, userID, challengeID int64) ([]*ChallengeSubmissionRecordResp, error) {
	if s.runtimeSubject == nil {
		return nil, errcode.ErrInternal.WithCause(errors.New("practice runtime subject repository is nil"))
	}
	challengeItem, err := s.runtimeSubject.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeChallengeNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	items, err := s.repo.ListChallengeSubmissions(ctx, userID, challengeID, 20)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*ChallengeSubmissionRecordResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, challengeSubmissionRecordRespFromModel(item))
	}
	return resp, nil
}

func challengeSubmissionRecordRespFromModel(item model.Submission) *ChallengeSubmissionRecordResp {
	status := SubmissionStatusIncorrect
	answer := ""

	if item.ReviewStatus == model.SubmissionReviewStatusPending {
		status = SubmissionStatusPendingReview
		answer = item.Flag
	} else if item.IsCorrect {
		status = SubmissionStatusCorrect
	}

	resp := practiceCommandResponseMapperInst.ToChallengeSubmissionRecordRespBasePtr(&item)
	resp.Status = status
	resp.Answer = answer
	return resp
}
