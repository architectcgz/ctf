package ops

import (
	"context"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

const (
	submitBurstWindow   = 30 * time.Minute
	loginSharedIPWindow = 24 * time.Hour
	submitBurstLimit    = 2000
	loginBurstLimit     = 2000
	submitBurstMinCount = 5
	sharedIPMinUsers    = 2
	maxRiskRows         = 10
)

type RiskService struct {
	repo *RiskRepository
	log  *zap.Logger
}

func NewRiskService(repo *RiskRepository, log *zap.Logger) *RiskService {
	if log == nil {
		log = zap.NewNop()
	}
	return &RiskService{repo: repo, log: log}
}

func (s *RiskService) GetCheatDetection(ctx context.Context) (*dto.CheatDetectionResp, error) {
	submitEvents, err := s.repo.ListRecentSubmitEvents(ctx, time.Now().Add(-submitBurstWindow), submitBurstLimit)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	loginEvents, err := s.repo.ListRecentLoginEvents(ctx, time.Now().Add(-loginSharedIPWindow), loginBurstLimit)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	suspects, suspectUserIDs := aggregateSubmitBursts(submitEvents)
	sharedIPs, sharedIPUserIDs := aggregateSharedIPs(loginEvents)

	affectedUsers := make(map[int64]struct{}, len(suspectUserIDs)+len(sharedIPUserIDs))
	for userID := range suspectUserIDs {
		affectedUsers[userID] = struct{}{}
	}
	for userID := range sharedIPUserIDs {
		affectedUsers[userID] = struct{}{}
	}

	return &dto.CheatDetectionResp{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Summary: dto.CheatDetectionSummary{
			SubmitBurstUsers: len(suspects),
			SharedIPGroups:   len(sharedIPs),
			AffectedUsers:    len(affectedUsers),
		},
		Suspects:  suspects,
		SharedIPs: sharedIPs,
	}, nil
}

func aggregateSubmitBursts(events []riskAuditEvent) ([]dto.CheatDetectionUser, map[int64]struct{}) {
	type item struct {
		username string
		count    int
		lastSeen time.Time
	}

	index := make(map[int64]*item)
	for _, event := range events {
		if event.UserID == nil || *event.UserID <= 0 {
			continue
		}
		entry, ok := index[*event.UserID]
		if !ok {
			entry = &item{username: event.Username}
			index[*event.UserID] = entry
		}
		entry.count++
		if event.CreatedAt.After(entry.lastSeen) {
			entry.lastSeen = event.CreatedAt
		}
	}

	userIDs := make(map[int64]struct{})
	rows := make([]dto.CheatDetectionUser, 0)
	for userID, entry := range index {
		if entry.count < submitBurstMinCount {
			continue
		}
		userIDs[userID] = struct{}{}
		rows = append(rows, dto.CheatDetectionUser{
			UserID:      userID,
			Username:    entry.username,
			SubmitCount: entry.count,
			LastSeenAt:  entry.lastSeen.Format(time.RFC3339),
			Reason:      "短时间内提交次数异常偏高",
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].SubmitCount == rows[j].SubmitCount {
			return rows[i].LastSeenAt > rows[j].LastSeenAt
		}
		return rows[i].SubmitCount > rows[j].SubmitCount
	})
	if len(rows) > maxRiskRows {
		rows = rows[:maxRiskRows]
	}
	return rows, userIDs
}

func aggregateSharedIPs(events []riskAuditEvent) ([]dto.CheatDetectionIPGroup, map[int64]struct{}) {
	type item struct {
		usernames map[string]struct{}
		userIDs   map[int64]struct{}
	}

	index := make(map[string]*item)
	for _, event := range events {
		ip := strings.TrimSpace(event.IPAddress)
		if ip == "" || event.UserID == nil || *event.UserID <= 0 {
			continue
		}
		entry, ok := index[ip]
		if !ok {
			entry = &item{
				usernames: make(map[string]struct{}),
				userIDs:   make(map[int64]struct{}),
			}
			index[ip] = entry
		}
		entry.userIDs[*event.UserID] = struct{}{}
		if username := strings.TrimSpace(event.Username); username != "" {
			entry.usernames[username] = struct{}{}
		}
	}

	affectedUsers := make(map[int64]struct{})
	rows := make([]dto.CheatDetectionIPGroup, 0)
	for ip, entry := range index {
		if len(entry.userIDs) < sharedIPMinUsers {
			continue
		}
		for userID := range entry.userIDs {
			affectedUsers[userID] = struct{}{}
		}
		usernames := make([]string, 0, len(entry.usernames))
		for username := range entry.usernames {
			usernames = append(usernames, username)
		}
		sort.Strings(usernames)
		rows = append(rows, dto.CheatDetectionIPGroup{
			IP:        ip,
			UserCount: len(entry.userIDs),
			Usernames: usernames,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].UserCount == rows[j].UserCount {
			return rows[i].IP < rows[j].IP
		}
		return rows[i].UserCount > rows[j].UserCount
	})
	if len(rows) > maxRiskRows {
		rows = rows[:maxRiskRows]
	}
	return rows, affectedUsers
}
