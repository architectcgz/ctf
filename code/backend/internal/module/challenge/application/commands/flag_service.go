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
	repo         ports.ChallengeRepository
	globalSecret string
}

func NewFlagService(repo ports.ChallengeRepository, globalSecret string) (*FlagService, error) {
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

	challenge.FlagType = model.FlagTypeDynamic
	if flagPrefix != "" {
		challenge.FlagPrefix = flagPrefix
	}
	return s.repo.Update(challenge)
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
