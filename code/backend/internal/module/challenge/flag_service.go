package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"os"

	"gorm.io/gorm"
)

type FlagService struct {
	db *gorm.DB
}

func NewFlagService(db *gorm.DB) *FlagService {
	return &FlagService{db: db}
}

// ConfigureStaticFlag 配置静态 Flag
func (s *FlagService) ConfigureStaticFlag(challengeID int64, flag string) error {
	var challenge model.Challenge
	if err := s.db.First(&challenge, challengeID).Error; err != nil {
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

	return s.db.Model(&challenge).Updates(map[string]interface{}{
		"flag_type": model.FlagTypeStatic,
		"flag_hash": hash,
		"flag_salt": salt,
	}).Error
}

// ConfigureDynamicFlag 配置动态 Flag
func (s *FlagService) ConfigureDynamicFlag(challengeID int64) error {
	var challenge model.Challenge
	if err := s.db.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrNotFound
		}
		return err
	}

	return s.db.Model(&challenge).Updates(map[string]interface{}{
		"flag_type": model.FlagTypeDynamic,
		"flag_hash": nil,
		"flag_salt": nil,
	}).Error
}

// GenerateDynamicFlag 生成动态 Flag
func (s *FlagService) GenerateDynamicFlag(userID, challengeID int64, nonce string) (string, error) {
	globalSecret := os.Getenv("CTF_FLAG_SECRET")
	if globalSecret == "" {
		return "", errcode.ErrInternal
	}

	return crypto.GenerateDynamicFlag(userID, challengeID, globalSecret, nonce), nil
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
		Configured: configured,
	}, nil
}
