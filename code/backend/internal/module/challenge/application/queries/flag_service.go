package queries

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type FlagService struct {
	repo         ports.ChallengeFlagRepository
	globalSecret string
}

func NewFlagService(repo ports.ChallengeFlagRepository, globalSecret string) (*FlagService, error) {
	secret := strings.TrimSpace(globalSecret)
	if secret == "" {
		return nil, fmt.Errorf("container.flag_global_secret 未配置")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("container.flag_global_secret 长度不足 32 字节，当前长度: %d", len(secret))
	}
	return &FlagService{
		repo:         repo,
		globalSecret: secret,
	}, nil
}

func (s *FlagService) GenerateDynamicFlag(ctx context.Context, userID, challengeID int64, nonce string) (string, error) {
	if nonce == "" {
		return "", errcode.ErrInvalidParams
	}

	challenge, err := s.loadChallenge(ctx, challengeID)
	if err != nil {
		return "", err
	}
	return crypto.GenerateDynamicFlag(userID, challengeID, s.globalSecret, nonce, challenge.FlagPrefix), nil
}

func (s *FlagService) ValidateFlag(ctx context.Context, userID, challengeID int64, input string, nonce string) (bool, error) {
	challenge, err := s.loadChallenge(ctx, challengeID)
	if err != nil {
		return false, err
	}

	switch challenge.FlagType {
	case model.FlagTypeStatic:
		hash := crypto.HashStaticFlag(input, challenge.FlagSalt)
		return crypto.ValidateFlag(hash, challenge.FlagHash), nil
	case model.FlagTypeRegex:
		return regexp.MatchString(challenge.FlagRegex, input)
	case model.FlagTypeManualReview:
		return false, nil
	case model.FlagTypeDynamic:
		expectedFlag, err := s.GenerateDynamicFlag(ctx, userID, challengeID, nonce)
		if err != nil {
			return false, err
		}
		return crypto.ValidateFlag(input, expectedFlag), nil
	default:
		return false, errcode.ErrInvalidParams.WithCause(fmt.Errorf("unsupported flag type %s", challenge.FlagType))
	}
}

func (s *FlagService) GetFlagConfig(ctx context.Context, challengeID int64) (*dto.FlagResp, error) {
	challenge, err := s.loadChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	configured := false
	if challenge.FlagType == model.FlagTypeStatic && challenge.FlagHash != "" {
		configured = true
	} else if challenge.FlagType == model.FlagTypeDynamic {
		configured = true
	} else if challenge.FlagType == model.FlagTypeRegex && strings.TrimSpace(challenge.FlagRegex) != "" {
		configured = true
	} else if challenge.FlagType == model.FlagTypeManualReview {
		configured = true
	}

	resp := challengeQueryResponseMapperInst.ToFlagRespBasePtr(challenge)
	resp.Configured = configured
	return resp, nil
}

func (s *FlagService) loadChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	challenge, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, ports.ErrChallengeFlagChallengeNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return challenge, nil
}
