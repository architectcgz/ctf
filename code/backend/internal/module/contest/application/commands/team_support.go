package commands

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) ensureApprovedRegistration(contestID, userID int64) error {
	registration, err := s.teamRepo.FindContestRegistration(contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotRegistered
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
		return err
	}
	return nil
}

func generateInviteCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := base32.StdEncoding.EncodeToString(bytes)
	code = strings.ReplaceAll(code, "=", "")
	if len(code) > 6 {
		code = code[:6]
	}
	return code, nil
}

func isUniqueConflict(err error) bool {
	if err == nil {
		return false
	}
	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "duplicate") || strings.Contains(lowered, "unique")
}

func teamHasMember(members []*model.TeamMember, userID int64) bool {
	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}
