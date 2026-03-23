package application

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type WriteupService struct {
	repo ChallengeRepository
}

func NewWriteupService(repo ChallengeRepository) *WriteupService {
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
	return adminWriteupResp(item), nil
}

func (s *WriteupService) GetAdmin(challengeID int64) (*dto.AdminChallengeWriteupResp, error) {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeID(challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return adminWriteupResp(item), nil
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

func (s *WriteupService) GetPublished(userID, challengeID int64) (*dto.ChallengeWriteupResp, error) {
	challengeItem, err := s.repo.FindByID(challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	item, err := s.repo.FindReleasedWriteupByChallengeID(challengeID, time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}

	isSolved, err := s.repo.GetSolvedStatus(userID, challengeID)
	if err != nil {
		isSolved = false
	}

	return &dto.ChallengeWriteupResp{
		ID:                     item.ID,
		ChallengeID:            item.ChallengeID,
		Title:                  item.Title,
		Content:                item.Content,
		Visibility:             item.Visibility,
		ReleaseAt:              item.ReleaseAt,
		IsReleased:             true,
		RequiresSpoilerWarning: !isSolved,
		CreatedAt:              item.CreatedAt,
		UpdatedAt:              item.UpdatedAt,
	}, nil
}

func adminWriteupResp(item *model.ChallengeWriteup) *dto.AdminChallengeWriteupResp {
	return &dto.AdminChallengeWriteupResp{
		ID:          item.ID,
		ChallengeID: item.ChallengeID,
		Title:       item.Title,
		Content:     item.Content,
		Visibility:  item.Visibility,
		ReleaseAt:   item.ReleaseAt,
		CreatedBy:   item.CreatedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
