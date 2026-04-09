package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

var flagPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+\{[^\{\}\n\r]+\}$`)

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

func (s *FlagService) ConfigureStaticFlag(challengeID int64, flag, flagPrefix string) error {
	if !flagPattern.MatchString(flag) {
		return errcode.New(10001, "Flag 格式错误，必须以 prefix{ 开头并以 } 结尾，如 flag{abc123}", 400)
	}
	if len(flag) > 256 {
		return errcode.New(10001, fmt.Sprintf("Flag 长度不能超过 256 字符，当前长度: %d", len(flag)), 400)
	}

	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return err
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	s.resetNonDynamicFlagFields(challenge)
	challenge.FlagType = model.FlagTypeStatic
	challenge.FlagHash = crypto.HashStaticFlag(flag, salt)
	challenge.FlagSalt = salt
	if flagPrefix != "" {
		challenge.FlagPrefix = flagPrefix
	}
	return s.repo.Update(challenge)
}

func (s *FlagService) ConfigureDynamicFlag(challengeID int64, flagPrefix string) error {
	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return err
	}
	if challenge.InstanceSharing == model.InstanceSharingShared {
		return errcode.ErrInvalidParams.WithCause(errors.New("共享实例策略不支持动态 Flag"))
	}

	s.resetNonDynamicFlagFields(challenge)
	challenge.FlagType = model.FlagTypeDynamic
	if flagPrefix != "" {
		challenge.FlagPrefix = flagPrefix
	}
	return s.repo.Update(challenge)
}

func (s *FlagService) ConfigureRegexFlag(challengeID int64, flagRegex, flagPrefix string) error {
	compiled, err := regexp.Compile(strings.TrimSpace(flagRegex))
	if err != nil {
		return errcode.New(10001, "Regex Flag 配置无效: "+err.Error(), 400)
	}

	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return err
	}

	s.resetNonDynamicFlagFields(challenge)
	challenge.FlagType = model.FlagTypeRegex
	challenge.FlagRegex = compiled.String()
	if flagPrefix != "" {
		challenge.FlagPrefix = flagPrefix
	}
	return s.repo.Update(challenge)
}

func (s *FlagService) ConfigureManualReviewFlag(challengeID int64) error {
	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return err
	}

	s.resetNonDynamicFlagFields(challenge)
	challenge.FlagType = model.FlagTypeManualReview
	return s.repo.Update(challenge)
}

func (s *FlagService) resetNonDynamicFlagFields(challenge *model.Challenge) {
	if challenge == nil {
		return
	}
	challenge.FlagHash = ""
	challenge.FlagSalt = ""
	challenge.FlagRegex = ""
}

func (s *FlagService) loadChallenge(challengeID int64) (*model.Challenge, error) {
	challenge, err := s.repo.FindByID(challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return challenge, nil
}
