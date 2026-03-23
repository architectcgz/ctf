package application

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

var flagPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+\{[^\{\}\n\r]+\}$`)

type FlagService struct {
	repo         ChallengeRepository
	globalSecret string
}

func NewFlagService(repo ChallengeRepository, globalSecret string) (*FlagService, error) {
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
	// 校验 Flag 格式
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

// ConfigureDynamicFlag 配置动态 Flag
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

func (s *FlagService) GenerateDynamicFlag(userID, challengeID int64, nonce string) (string, error) {
	if nonce == "" {
		return "", errcode.ErrInvalidParams
	}

	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return "", err
	}

	return crypto.GenerateDynamicFlag(userID, challengeID, s.globalSecret, nonce, challenge.FlagPrefix), nil
}

// ValidateFlag 验证 Flag
func (s *FlagService) ValidateFlag(userID, challengeID int64, input string, nonce string) (bool, error) {
	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return false, err
	}

	if challenge.FlagType == model.FlagTypeStatic {
		hash := crypto.HashStaticFlag(input, challenge.FlagSalt)
		return crypto.ValidateFlag(hash, challenge.FlagHash), nil
	}

	expectedFlag, err := s.GenerateDynamicFlag(userID, challengeID, nonce)
	if err != nil {
		return false, err
	}

	return crypto.ValidateFlag(input, expectedFlag), nil
}

// GetFlagConfig 获取 Flag 配置
func (s *FlagService) GetFlagConfig(challengeID int64) (*dto.FlagResp, error) {
	challenge, err := s.loadChallenge(challengeID)
	if err != nil {
		return nil, err
	}

	configured := false
	if challenge.FlagType == model.FlagTypeStatic && challenge.FlagHash != "" {
		configured = true
	} else if challenge.FlagType == model.FlagTypeDynamic {
		configured = true
	}

	return &dto.FlagResp{
		FlagType:   challenge.FlagType,
		FlagPrefix: challenge.FlagPrefix,
		Configured: configured,
	}, nil
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
