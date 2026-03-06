package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"fmt"
	"os"
	"regexp"

	"gorm.io/gorm"
)

var flagPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+\{[^\{\}\n\r]+\}$`)

type FlagService struct {
	db           *gorm.DB
	globalSecret string
}

func NewFlagService(db *gorm.DB) (*FlagService, error) {
	secret := os.Getenv("CTF_FLAG_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("CTF_FLAG_SECRET 环境变量未配置")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("CTF_FLAG_SECRET 长度不足 32 字节，当前长度: %d", len(secret))
	}
	return &FlagService{
		db:           db,
		globalSecret: secret,
	}, nil
}

// ConfigureStaticFlag 配置静态 Flag
func (s *FlagService) ConfigureStaticFlag(challengeID int64, flag, flagPrefix string) error {
	// 校验 Flag 格式
	if !flagPattern.MatchString(flag) {
		return errcode.New(10001, "Flag 格式错误，必须以 prefix{ 开头并以 } 结尾，如 flag{abc123}", 400)
	}
	if len(flag) > 256 {
		return errcode.New(10001, fmt.Sprintf("Flag 长度不能超过 256 字符，当前长度: %d", len(flag)), 400)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var challenge model.Challenge
		if err := tx.First(&challenge, challengeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errcode.ErrNotFound
			}
			return err
		}

		salt, err := crypto.GenerateSalt()
		if err != nil {
			return err
		}

		hash := crypto.HashStaticFlag(flag, salt)

		updates := map[string]interface{}{
			"flag_type": model.FlagTypeStatic,
			"flag_hash": hash,
			"flag_salt": salt,
		}
		if flagPrefix != "" {
			updates["flag_prefix"] = flagPrefix
		}

		return tx.Model(&challenge).Updates(updates).Error
	})
}

// ConfigureDynamicFlag 配置动态 Flag
func (s *FlagService) ConfigureDynamicFlag(challengeID int64, flagPrefix string) error {
	var challenge model.Challenge
	if err := s.db.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrNotFound
		}
		return err
	}

	updates := map[string]interface{}{
		"flag_type": model.FlagTypeDynamic,
	}
	if flagPrefix != "" {
		updates["flag_prefix"] = flagPrefix
	}

	return s.db.Model(&challenge).Updates(updates).Error
}

// GenerateDynamicFlag 生成动态 Flag
// nonce 参数应从 instances.nonce 字段获取，由实例创建时生成
func (s *FlagService) GenerateDynamicFlag(userID, challengeID int64, nonce string) (string, error) {
	if nonce == "" {
		return "", errcode.ErrInvalidParams
	}

	var challenge model.Challenge
	if err := s.db.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errcode.ErrNotFound
		}
		return "", err
	}

	return crypto.GenerateDynamicFlag(userID, challengeID, s.globalSecret, nonce, challenge.FlagPrefix), nil
}

// ValidateFlag 验证 Flag
func (s *FlagService) ValidateFlag(userID, challengeID int64, input string, nonce string) (bool, error) {
	var challenge model.Challenge
	if err := s.db.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errcode.ErrNotFound
		}
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
	var challenge model.Challenge
	if err := s.db.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrNotFound
		}
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
