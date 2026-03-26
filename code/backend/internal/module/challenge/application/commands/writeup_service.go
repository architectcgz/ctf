package commands

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type WriteupService struct {
	repo challengeports.ChallengeWriteupRepository
}

func NewWriteupService(repo challengeports.ChallengeWriteupRepository) *WriteupService {
	return &WriteupService{repo: repo}
}

func (s *WriteupService) Upsert(challengeID, actorUserID int64, req *dto.UpsertChallengeWriteupReq) (*dto.AdminChallengeWriteupResp, error) {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if req.Visibility == model.WriteupVisibilityScheduled && req.ReleaseAt == nil {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("定时公开必须提供 release_at"))
	}

	writeup := &model.ChallengeWriteup{
		ChallengeID: challengeID,
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		Visibility:  req.Visibility,
		ReleaseAt:   req.ReleaseAt,
		CreatedBy:   &actorUserID,
		UpdatedAt:   time.Now(),
	}
	if writeup.Visibility != model.WriteupVisibilityScheduled {
		writeup.ReleaseAt = nil
	}
	if err := s.repo.UpsertWriteup(writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	return domain.AdminWriteupRespFromModel(item), nil
}

func (s *WriteupService) Delete(challengeID int64) error {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteWriteupByChallengeID(challengeID)
}
