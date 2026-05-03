package queries

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	readmodelports "ctf-platform/internal/module/teaching_readmodel/ports"
	"ctf-platform/pkg/errcode"
)

type QueryService struct {
	repo                  teachingReadModelQueryRepository
	recommendationService assessmentcontracts.RecommendationProvider
	pagination            config.PaginationConfig
	logger                *zap.Logger
}

type teachingReadModelQueryRepository interface {
	readmodelports.TeachingUserLookupRepository
	readmodelports.TeachingClassQueryRepository
	readmodelports.TeachingStudentDirectoryRepository
	readmodelports.TeachingStudentProfileRepository
	readmodelports.TeachingStudentActivityRepository
	readmodelports.TeachingClassInsightRepository
}

var _ Service = (*QueryService)(nil)

func NewQueryService(
	repo teachingReadModelQueryRepository,
	recommendationService assessmentcontracts.RecommendationProvider,
	pagination config.PaginationConfig,
	logger *zap.Logger,
) *QueryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QueryService{
		repo:                  repo,
		recommendationService: recommendationService,
		pagination:            pagination,
		logger:                logger,
	}
}

func (s *QueryService) ListClasses(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherClassQuery,
) ([]dto.TeacherClassItem, int64, int, int, error) {
	page, size := s.normalizeClassPagination(query)

	requester, err := s.repo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return nil, 0, 0, 0, errcode.ErrUnauthorized
	}

	if requesterRole == model.RoleAdmin {
		total, err := s.repo.CountClasses(ctx)
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
		}
		if total == 0 {
			return []dto.TeacherClassItem{}, 0, page, size, nil
		}

		items, err := s.repo.ListClasses(ctx, (page-1)*size, size)
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
		}
		return toClassItems(items), total, page, size, nil
	}

	className := strings.TrimSpace(requester.ClassName)
	if className == "" {
		return []dto.TeacherClassItem{}, 0, page, size, nil
	}

	count, err := s.repo.CountStudentsByClass(ctx, className)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}

	if (page-1)*size >= 1 {
		return []dto.TeacherClassItem{}, 1, page, size, nil
	}

	return []dto.TeacherClassItem{{
		Name:         className,
		StudentCount: count,
	}}, 1, page, size, nil
}

func (s *QueryService) normalizeClassPagination(query *dto.TeacherClassQuery) (int, int) {
	page := 1
	size := s.pagination.DefaultPageSize

	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.Size > 0 {
			size = query.Size
		}
	}

	if size < 1 {
		size = 20
	}
	if s.pagination.MaxPageSize > 0 && size > s.pagination.MaxPageSize {
		size = s.pagination.MaxPageSize
	}

	return page, size
}

func (s *QueryService) ListStudents(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherStudentDirectoryQuery,
) ([]dto.TeacherStudentItem, int64, int, int, error) {
	page, size := s.normalizeStudentPagination(query)

	var requester *model.User
	if requesterRole != model.RoleAdmin {
		var err error
		requester, err = s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return nil, 0, 0, 0, errcode.ErrUnauthorized
		}
	}

	className := ""
	keyword := ""
	studentNo := ""
	sortKey := "solved_count"
	sortOrder := "desc"
	if query != nil {
		className = strings.TrimSpace(query.ClassName)
		keyword = strings.TrimSpace(query.Keyword)
		studentNo = strings.TrimSpace(query.StudentNo)
		if strings.TrimSpace(query.SortKey) != "" {
			sortKey = strings.TrimSpace(query.SortKey)
		}
		if strings.TrimSpace(query.SortOrder) != "" {
			sortOrder = strings.TrimSpace(query.SortOrder)
		}
	}

	if requesterRole != model.RoleAdmin {
		requesterClassName := strings.TrimSpace(requester.ClassName)
		if requesterClassName == "" {
			return []dto.TeacherStudentItem{}, 0, page, size, nil
		}
		if className == "" {
			className = requesterClassName
		} else if className != requesterClassName {
			return nil, 0, 0, 0, errcode.ErrForbidden
		}
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	items, total, err := s.repo.ListStudents(ctx, className, keyword, studentNo, sortKey, sortOrder, startOfDay, (page-1)*size, size)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}
	return toStudentItems(items), total, page, size, nil
}

func (s *QueryService) normalizeStudentPagination(query *dto.TeacherStudentDirectoryQuery) (int, int) {
	page := 1
	size := s.pagination.DefaultPageSize

	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.Size > 0 {
			size = query.Size
		}
	}

	if size < 1 {
		size = 20
	}
	if s.pagination.MaxPageSize > 0 && size > s.pagination.MaxPageSize {
		size = s.pagination.MaxPageSize
	}

	return page, size
}

func (s *QueryService) ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherStudentQuery) ([]dto.TeacherStudentItem, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}

	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	studentNo := ""
	keyword := ""
	if query != nil {
		studentNo = strings.TrimSpace(query.StudentNo)
		keyword = strings.TrimSpace(query.Keyword)
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	items, err := s.repo.ListStudentsByClass(ctx, normalized, keyword, studentNo, startOfDay)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toStudentItems(items), nil
}

func (s *QueryService) GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassSummaryResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	summary, err := s.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toClassSummary(summary), nil
}

func (s *QueryService) GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassTrendResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	trend, err := s.repo.GetClassTrend(ctx, normalized, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toClassTrend(trend), nil
}

func (s *QueryService) GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassReviewResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())

	students, err := s.repo.ListStudentsByClass(ctx, normalized, "", "", startOfDay)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	summary, err := s.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	trend, err := s.repo.GetClassTrend(ctx, normalized, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	studentItems := toStudentItems(students)
	summaryDTO := toClassSummary(summary)
	trendDTO := toClassTrend(trend)

	items := make([]dto.TeacherClassReviewItem, 0, 4)
	riskStudents := selectRiskStudents(studentItems, 3)
	activeRate := summaryDTO.ActiveRate

	switch {
	case activeRate < 50:
		items = append(items, dto.TeacherClassReviewItem{
			Key:    "activity",
			Title:  "班级活跃度偏低",
			Detail: fmt.Sprintf("%s 近 7 天活跃率只有 %.0f%%，建议优先跟进 %d 名低活跃学生。", normalized, activeRate, len(riskStudents)),
			Accent: "danger",
		})
	case activeRate < 75:
		items = append(items, dto.TeacherClassReviewItem{
			Key:    "activity",
			Title:  "班级活跃度需要补强",
			Detail: fmt.Sprintf("%s 近 7 天活跃率为 %.0f%%，适合通过定向训练把低活跃学生重新拉回训练节奏。", normalized, activeRate),
			Accent: "warning",
		})
	default:
		items = append(items, dto.TeacherClassReviewItem{
			Key:    "activity",
			Title:  "班级训练节奏整体稳定",
			Detail: fmt.Sprintf("%s 近 7 天活跃率达到 %.0f%%，当前更适合做薄弱维度补强而不是全面催学。", normalized, activeRate),
			Accent: "success",
		})
	}

	if weakDimension, weakStudents := selectWeakDimensionStudents(studentItems); weakDimension != "" {
		item := dto.TeacherClassReviewItem{
			Key:      "weak_dimension",
			Title:    "优先补薄弱维度",
			Detail:   fmt.Sprintf("%s 是当前最集中的薄弱项，涉及 %d 名学生，建议本周统一布置该维度基础题。", weakDimension, len(weakStudents)),
			Accent:   "primary",
			Students: toReviewStudentRefs(limitStudents(weakStudents, 3)),
		}
		if len(weakStudents) >= 3 {
			item.Accent = "warning"
		}
		if recommendation := s.firstStudentRecommendation(ctx, weakStudents, 1); recommendation != nil {
			item.Recommendation = recommendation
		}
		items = append(items, item)
	}

	if len(riskStudents) > 0 {
		item := dto.TeacherClassReviewItem{
			Key:      "focus_students",
			Title:    "先跟进重点学生",
			Detail:   fmt.Sprintf("建议教师先跟进 %s，并优先布置推荐题做补强训练。", joinStudentNames(riskStudents)),
			Accent:   "primary",
			Students: toReviewStudentRefs(riskStudents),
		}
		if recommendation := s.firstStudentRecommendation(ctx, riskStudents, 1); recommendation != nil {
			item.Recommendation = recommendation
		}
		items = append(items, item)
	}

	if trendItem := buildTrendReviewItem(trendDTO); trendItem != nil {
		items = append(items, *trendItem)
	}

	return &dto.TeacherClassReviewResp{
		ClassName: normalized,
		Items:     items,
	}, nil
}

func (s *QueryService) GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.TeacherProgressResp, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	totalChallenges, err := s.repo.CountPublishedChallenges(ctx)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	solvedChallenges, err := s.repo.CountSolvedChallenges(ctx, student.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	categoryRows, err := s.repo.GetCategoryProgress(ctx, student.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	difficultyRows, err := s.repo.GetDifficultyProgress(ctx, student.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TeacherProgressResp{
		TotalChallenges:  int(totalChallenges),
		SolvedChallenges: int(solvedChallenges),
		ByCategory:       toProgressBreakdownMap(categoryRows),
		ByDifficulty:     toProgressBreakdownMap(difficultyRows),
	}, nil
}

func (s *QueryService) GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) ([]dto.TeacherRecommendationItem, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	result, err := s.recommendationService.Recommend(ctx, student.ID, limit)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]dto.TeacherRecommendationItem, 0, len(result.Challenges))
	for _, challenge := range result.Challenges {
		items = append(items, dto.TeacherRecommendationItem{
			ChallengeID: challenge.ID,
			Title:       challenge.Title,
			Category:    challenge.Category,
			Difficulty:  challenge.Difficulty,
			Reason:      challenge.Reason,
		})
	}

	return items, nil
}

func (s *QueryService) GetStudentTimeline(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit, offset int) (*dto.TimelineResp, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.GetStudentTimeline(ctx, student.ID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TimelineResp{Events: toTimelineEvents(events)}, nil
}

func (s *QueryService) GetStudentEvidence(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *dto.TeacherEvidenceQuery) (*dto.TeacherEvidenceResp, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	repoQuery := buildEvidenceQuery(query)
	events, err := s.repo.GetStudentEvidence(ctx, student.ID, repoQuery)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	events = filterEvidenceEvents(events, repoQuery)
	events = paginateEvidenceEvents(events, repoQuery)

	resp := &dto.TeacherEvidenceResp{
		Events: make([]dto.TeacherEvidenceEvent, 0, len(events)),
	}
	for _, event := range events {
		resp.Events = append(resp.Events, dto.TeacherEvidenceEvent{
			Type:        event.Type,
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Detail:      event.Detail,
			Timestamp:   event.Timestamp,
			Meta:        event.Meta,
		})
		resp.Summary.TotalEvents++
		switch event.Type {
		case "instance_proxy_request":
			resp.Summary.ProxyRequestCount++
		case "challenge_submission", "awd_attack_submission":
			resp.Summary.SubmitCount++
			if evidenceEventSucceeded(event) {
				resp.Summary.SuccessCount++
			}
		}
		if resp.Summary.ChallengeID == 0 {
			resp.Summary.ChallengeID = event.ChallengeID
		}
	}
	if repoQuery.ChallengeID != nil {
		resp.Summary.ChallengeID = *repoQuery.ChallengeID
	}

	return resp, nil
}

func (s *QueryService) GetStudentAttackSessions(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *dto.TeacherAttackSessionQuery) (*dto.TeacherAttackSessionResp, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	evidenceQuery := buildEvidenceQuery(&dto.TeacherEvidenceQuery{
		ChallengeID: query.ChallengeID,
		ContestID:   query.ContestID,
		RoundID:     query.RoundID,
		Limit:       0,
		Offset:      0,
	})
	events, err := s.repo.GetStudentEvidence(ctx, student.ID, evidenceQuery)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	events = filterEvidenceEvents(events, evidenceQuery)
	events = filterAttackSessionEvents(events, query)

	sessions := buildAttackSessions(student.ID, events)
	if query != nil && query.Result != "" {
		filtered := sessions[:0]
		for _, session := range sessions {
			if session.Result == query.Result {
				filtered = append(filtered, session)
			}
		}
		sessions = filtered
	}

	resp := &dto.TeacherAttackSessionResp{
		Summary:  summarizeAttackSessions(sessions),
		Sessions: paginateAttackSessions(sessions, query),
	}
	if query != nil && query.WithEvents != nil && !*query.WithEvents {
		for index := range resp.Sessions {
			resp.Sessions[index].Events = nil
		}
	}
	return resp, nil
}

const attackSessionGap = time.Hour

func buildEvidenceQuery(query *dto.TeacherEvidenceQuery) readmodelports.EvidenceQuery {
	if query == nil {
		return readmodelports.EvidenceQuery{}
	}
	return readmodelports.EvidenceQuery{
		ChallengeID: query.ChallengeID,
		ContestID:   query.ContestID,
		RoundID:     query.RoundID,
		EventType:   strings.TrimSpace(query.EventType),
		From:        normalizeUTC(query.From),
		To:          normalizeUTC(query.To),
		Limit:       query.Limit,
		Offset:      query.Offset,
	}
}

func normalizeUTC(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

func filterEvidenceEvents(events []readmodelports.EvidenceEventRecord, query readmodelports.EvidenceQuery) []readmodelports.EvidenceEventRecord {
	if len(events) == 0 {
		return events
	}
	filtered := make([]readmodelports.EvidenceEventRecord, 0, len(events))
	for _, event := range events {
		if query.ContestID != nil && !int64PtrEqual(event.ContestID, *query.ContestID) {
			continue
		}
		if query.RoundID != nil && !int64PtrEqual(event.RoundID, *query.RoundID) {
			continue
		}
		if query.EventType != "" && event.Type != query.EventType {
			continue
		}
		if query.From != nil && event.Timestamp.Before(*query.From) {
			continue
		}
		if query.To != nil && event.Timestamp.After(*query.To) {
			continue
		}
		filtered = append(filtered, event)
	}
	return filtered
}

func paginateEvidenceEvents(events []readmodelports.EvidenceEventRecord, query readmodelports.EvidenceQuery) []readmodelports.EvidenceEventRecord {
	if len(events) == 0 {
		return []readmodelports.EvidenceEventRecord{}
	}
	offset := query.Offset
	if offset >= len(events) {
		return []readmodelports.EvidenceEventRecord{}
	}
	if offset < 0 {
		offset = 0
	}
	if query.Limit <= 0 {
		return events[offset:]
	}
	end := offset + query.Limit
	if end > len(events) {
		end = len(events)
	}
	return events[offset:end]
}

func filterAttackSessionEvents(events []readmodelports.EvidenceEventRecord, query *dto.TeacherAttackSessionQuery) []readmodelports.EvidenceEventRecord {
	if query == nil {
		return events
	}
	filtered := make([]readmodelports.EvidenceEventRecord, 0, len(events))
	for _, event := range events {
		if query.Mode != "" && attackEventMode(event) != query.Mode {
			continue
		}
		if query.ContestID != nil && !int64PtrEqual(event.ContestID, *query.ContestID) {
			continue
		}
		if query.RoundID != nil && !int64PtrEqual(event.RoundID, *query.RoundID) {
			continue
		}
		filtered = append(filtered, event)
	}
	return filtered
}

func buildAttackSessions(studentID int64, events []readmodelports.EvidenceEventRecord) []dto.TeacherAttackSession {
	if len(events) == 0 {
		return []dto.TeacherAttackSession{}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	grouped := make(map[string][]readmodelports.EvidenceEventRecord)
	order := make([]string, 0)
	for _, event := range events {
		key := attackSessionGroupKey(event)
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], event)
	}

	sessions := make([]dto.TeacherAttackSession, 0, len(grouped))
	for _, key := range order {
		chunk := make([]readmodelports.EvidenceEventRecord, 0)
		for _, event := range grouped[key] {
			if len(chunk) > 0 && event.Timestamp.Sub(chunk[len(chunk)-1].Timestamp) > attackSessionGap {
				sessions = append(sessions, buildAttackSession(studentID, len(sessions)+1, chunk))
				chunk = chunk[:0]
			}
			chunk = append(chunk, event)
		}
		if len(chunk) > 0 {
			sessions = append(sessions, buildAttackSession(studentID, len(sessions)+1, chunk))
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})
	return sessions
}

func buildAttackSession(studentID int64, sequence int, events []readmodelports.EvidenceEventRecord) dto.TeacherAttackSession {
	first := events[0]
	last := events[len(events)-1]
	sessionID := fmt.Sprintf("sess_%d_%d", studentID, sequence)
	session := dto.TeacherAttackSession{
		ID:           sessionID,
		Mode:         attackEventMode(first),
		StudentID:    studentID,
		TeamID:       first.TeamID,
		ChallengeID:  attackEventChallengeID(first),
		ContestID:    first.ContestID,
		RoundID:      first.RoundID,
		ServiceID:    first.ServiceID,
		VictimTeamID: first.VictimTeamID,
		Title:        first.Title,
		StartedAt:    first.Timestamp,
		EndedAt:      last.Timestamp,
		Result:       deriveAttackSessionResult(events),
		EventCount:   len(events),
		Events:       make([]dto.TeacherAttackEvent, 0, len(events)),
	}
	for index, event := range events {
		attackEvent := toAttackEvent(studentID, sessionID, index, event)
		if attackEvent.CaptureAvailable {
			session.CaptureCount++
		}
		session.Events = append(session.Events, attackEvent)
	}
	return session
}

func toAttackEvent(studentID int64, sessionID string, index int, event readmodelports.EvidenceEventRecord) dto.TeacherAttackEvent {
	return dto.TeacherAttackEvent{
		ID:         fmt.Sprintf("%s_evt_%d", sessionID, index+1),
		SessionID:  sessionID,
		Type:       event.Type,
		Stage:      evidenceEventStage(event),
		Source:     evidenceEventSource(event),
		OccurredAt: event.Timestamp,
		Actor: dto.TeacherAttackActor{
			UserID: studentID,
			TeamID: event.TeamID,
		},
		Target: dto.TeacherAttackTarget{
			ChallengeID:  attackEventChallengeID(event),
			ContestID:    event.ContestID,
			RoundID:      event.RoundID,
			ServiceID:    event.ServiceID,
			VictimTeamID: event.VictimTeamID,
		},
		Summary:          event.Detail,
		Meta:             event.Meta,
		CaptureAvailable: false,
	}
}

func summarizeAttackSessions(sessions []dto.TeacherAttackSession) dto.TeacherAttackSessionSummary {
	summary := dto.TeacherAttackSessionSummary{TotalSessions: len(sessions)}
	for _, session := range sessions {
		summary.EventCount += session.EventCount
		summary.CaptureAvailableCount += session.CaptureCount
		switch session.Result {
		case "success":
			summary.SuccessCount++
		case "failed":
			summary.FailedCount++
		case "in_progress":
			summary.InProgressCount++
		default:
			summary.UnknownCount++
		}
	}
	return summary
}

func paginateAttackSessions(sessions []dto.TeacherAttackSession, query *dto.TeacherAttackSessionQuery) []dto.TeacherAttackSession {
	limit := 20
	offset := 0
	if query != nil {
		if query.Limit > 0 {
			limit = query.Limit
		}
		if query.Offset > 0 {
			offset = query.Offset
		}
	}
	if offset >= len(sessions) {
		return []dto.TeacherAttackSession{}
	}
	end := offset + limit
	if end > len(sessions) {
		end = len(sessions)
	}
	return sessions[offset:end]
}

func attackSessionGroupKey(event readmodelports.EvidenceEventRecord) string {
	if attackEventMode(event) == "awd" {
		return fmt.Sprintf("awd:%s:%s:%s:%s", ptrKey(event.TeamID), ptrKey(event.ContestID), ptrKey(event.ServiceID), ptrKey(event.VictimTeamID))
	}
	return fmt.Sprintf("%s:%d:%s", attackEventMode(event), event.ChallengeID, ptrKey(event.ContestID))
}

func attackEventMode(event readmodelports.EvidenceEventRecord) string {
	if strings.HasPrefix(event.Type, "awd_") {
		return "awd"
	}
	if event.ContestID != nil {
		return "jeopardy"
	}
	return "practice"
}

func attackEventChallengeID(event readmodelports.EvidenceEventRecord) *int64 {
	if event.ChallengeID <= 0 {
		return nil
	}
	value := event.ChallengeID
	return &value
}

func deriveAttackSessionResult(events []readmodelports.EvidenceEventRecord) string {
	hasFailure := false
	hasAction := false
	for _, event := range events {
		switch event.Type {
		case "challenge_submission", "awd_attack_submission":
			if evidenceEventSucceeded(event) {
				return "success"
			}
			hasFailure = true
		case "instance_access", "instance_proxy_request", "awd_traffic":
			hasAction = true
		}
	}
	if hasFailure {
		return "failed"
	}
	if hasAction {
		return "in_progress"
	}
	return "unknown"
}

func evidenceEventSucceeded(event readmodelports.EvidenceEventRecord) bool {
	if event.Meta == nil {
		return false
	}
	if isCorrect, ok := event.Meta["is_correct"].(bool); ok && isCorrect {
		return true
	}
	if isSuccess, ok := event.Meta["is_success"].(bool); ok && isSuccess {
		return true
	}
	return false
}

func evidenceEventStage(event readmodelports.EvidenceEventRecord) string {
	if event.Stage != "" {
		return event.Stage
	}
	if event.Meta != nil {
		if stage, ok := event.Meta["event_stage"].(string); ok && stage != "" {
			return stage
		}
	}
	return "trace"
}

func evidenceEventSource(event readmodelports.EvidenceEventRecord) string {
	if event.Source != "" {
		return event.Source
	}
	switch event.Type {
	case "instance_access", "instance_proxy_request":
		return "audit_logs"
	case "challenge_submission", "manual_review":
		return "submissions"
	case "writeup":
		return "submission_writeups"
	case "awd_attack_submission":
		return "awd_attack_logs"
	case "awd_traffic":
		return "awd_traffic_events"
	default:
		return "unknown"
	}
}

func ptrKey(value *int64) string {
	if value == nil {
		return "0"
	}
	return fmt.Sprintf("%d", *value)
}

func int64PtrEqual(value *int64, expected int64) bool {
	return value != nil && *value == expected
}

func (s *QueryService) getAccessibleStudent(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*model.User, error) {
	student, err := s.repo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != model.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, student.ClassName); err != nil {
		return nil, err
	}
	return student, nil
}

func (s *QueryService) ensureClassAccess(ctx context.Context, requesterID int64, requesterRole, className string) error {
	if requesterRole == model.RoleAdmin {
		return nil
	}

	requester, err := s.repo.FindUserByID(ctx, requesterID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return errcode.ErrUnauthorized
	}

	if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != className {
		return errcode.ErrForbidden
	}
	return nil
}

func toClassItems(items []readmodelports.ClassItem) []dto.TeacherClassItem {
	if len(items) == 0 {
		return []dto.TeacherClassItem{}
	}

	result := make([]dto.TeacherClassItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.TeacherClassItem{
			Name:         item.Name,
			StudentCount: item.StudentCount,
		})
	}
	return result
}

func toStudentItems(items []readmodelports.StudentItem) []dto.TeacherStudentItem {
	if len(items) == 0 {
		return []dto.TeacherStudentItem{}
	}

	result := make([]dto.TeacherStudentItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.TeacherStudentItem{
			ID:               item.ID,
			Username:         item.Username,
			StudentNo:        item.StudentNo,
			Name:             item.Name,
			ClassName:        item.ClassName,
			SolvedCount:      item.SolvedCount,
			TotalScore:       item.TotalScore,
			RecentEventCount: item.RecentEventCount,
			WeakDimension:    item.WeakDimension,
		})
	}
	return result
}

func toClassSummary(summary *readmodelports.ClassSummary) *dto.TeacherClassSummaryResp {
	if summary == nil {
		return nil
	}

	return &dto.TeacherClassSummaryResp{
		ClassName:          summary.ClassName,
		StudentCount:       summary.StudentCount,
		AverageSolved:      summary.AverageSolved,
		ActiveStudentCount: summary.ActiveStudentCount,
		ActiveRate:         summary.ActiveRate,
		RecentEventCount:   summary.RecentEventCount,
	}
}

func toClassTrend(trend *readmodelports.ClassTrend) *dto.TeacherClassTrendResp {
	if trend == nil {
		return nil
	}

	points := make([]dto.TeacherClassTrendPoint, 0, len(trend.Points))
	for _, point := range trend.Points {
		points = append(points, dto.TeacherClassTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}

	return &dto.TeacherClassTrendResp{
		ClassName: trend.ClassName,
		Points:    points,
	}
}

func toTimelineEvents(events []readmodelports.TimelineEventRecord) []dto.TimelineEvent {
	if len(events) == 0 {
		return []dto.TimelineEvent{}
	}

	result := make([]dto.TimelineEvent, 0, len(events))
	for _, event := range events {
		result = append(result, dto.TimelineEvent{
			Type:        event.Type,
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Timestamp:   event.Timestamp,
			IsCorrect:   event.IsCorrect,
			Points:      event.Points,
			Detail:      event.Detail,
		})
	}
	return result
}

func toProgressBreakdownMap(rows []readmodelports.ProgressRow) map[string]dto.ProgressBreakdown {
	if len(rows) == 0 {
		return map[string]dto.ProgressBreakdown{}
	}

	result := make(map[string]dto.ProgressBreakdown, len(rows))
	for _, row := range rows {
		result[row.Key] = dto.ProgressBreakdown{
			Total:  row.Total,
			Solved: row.Solved,
		}
	}
	return result
}

func selectRiskStudents(students []dto.TeacherStudentItem, limit int) []dto.TeacherStudentItem {
	filtered := make([]dto.TeacherStudentItem, 0, len(students))
	for _, student := range students {
		if student.RecentEventCount <= 1 || student.SolvedCount <= 1 {
			filtered = append(filtered, student)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].RecentEventCount != filtered[j].RecentEventCount {
			return filtered[i].RecentEventCount < filtered[j].RecentEventCount
		}
		if filtered[i].SolvedCount != filtered[j].SolvedCount {
			return filtered[i].SolvedCount < filtered[j].SolvedCount
		}
		return filtered[i].Username < filtered[j].Username
	})

	return limitStudents(filtered, limit)
}

func selectWeakDimensionStudents(students []dto.TeacherStudentItem) (string, []dto.TeacherStudentItem) {
	counter := make(map[string]int)
	grouped := make(map[string][]dto.TeacherStudentItem)
	for _, student := range students {
		if student.WeakDimension == nil {
			continue
		}
		key := strings.TrimSpace(*student.WeakDimension)
		if key == "" {
			continue
		}
		counter[key]++
		grouped[key] = append(grouped[key], student)
	}

	bestDimension := ""
	bestCount := 0
	for dimension, count := range counter {
		if count > bestCount || (count == bestCount && (bestDimension == "" || dimension < bestDimension)) {
			bestDimension = dimension
			bestCount = count
		}
	}
	if bestDimension == "" {
		return "", nil
	}

	studentsInDimension := grouped[bestDimension]
	sort.Slice(studentsInDimension, func(i, j int) bool {
		if studentsInDimension[i].SolvedCount != studentsInDimension[j].SolvedCount {
			return studentsInDimension[i].SolvedCount < studentsInDimension[j].SolvedCount
		}
		if studentsInDimension[i].RecentEventCount != studentsInDimension[j].RecentEventCount {
			return studentsInDimension[i].RecentEventCount < studentsInDimension[j].RecentEventCount
		}
		return studentsInDimension[i].Username < studentsInDimension[j].Username
	})
	return bestDimension, studentsInDimension
}

func limitStudents(students []dto.TeacherStudentItem, limit int) []dto.TeacherStudentItem {
	if limit <= 0 || len(students) <= limit {
		return students
	}
	return students[:limit]
}

func toReviewStudentRefs(students []dto.TeacherStudentItem) []dto.TeacherReviewStudentRef {
	items := make([]dto.TeacherReviewStudentRef, 0, len(students))
	for _, student := range students {
		items = append(items, dto.TeacherReviewStudentRef{
			ID:       student.ID,
			Username: student.Username,
			Name:     student.Name,
		})
	}
	return items
}

func joinStudentNames(students []dto.TeacherStudentItem) string {
	names := make([]string, 0, len(students))
	for _, student := range students {
		if student.Name != nil && strings.TrimSpace(*student.Name) != "" {
			names = append(names, strings.TrimSpace(*student.Name))
			continue
		}
		names = append(names, student.Username)
	}
	return strings.Join(names, "、")
}

func buildTrendReviewItem(trend *dto.TeacherClassTrendResp) *dto.TeacherClassReviewItem {
	if trend == nil || len(trend.Points) < 2 {
		return nil
	}
	first := trend.Points[0]
	last := trend.Points[len(trend.Points)-1]
	eventDelta := last.EventCount - first.EventCount
	solveDelta := last.SolveCount - first.SolveCount

	item := &dto.TeacherClassReviewItem{
		Key:    "trend",
		Title:  "观察最近一周走势",
		Accent: "success",
		Detail: fmt.Sprintf("最近一周训练事件变化 %d，成功解题变化 %d，说明班级训练投入仍在向前推进。", eventDelta, solveDelta),
	}
	if solveDelta < 0 || eventDelta < 0 {
		item.Accent = "warning"
		item.Detail = fmt.Sprintf("最近一周训练事件变化 %d，成功解题变化 %d，需要关注训练投入是否正在下滑。", eventDelta, solveDelta)
	}
	return item
}

func (s *QueryService) firstStudentRecommendation(ctx context.Context, students []dto.TeacherStudentItem, limit int) *dto.TeacherRecommendationItem {
	if s.recommendationService == nil {
		return nil
	}
	for _, student := range students {
		result, err := s.recommendationService.Recommend(ctx, student.ID, limit)
		if err != nil {
			s.logger.Warn("recommend_student_for_class_review_failed", zap.Int64("student_id", student.ID), zap.Error(err))
			continue
		}
		if result == nil || len(result.Challenges) == 0 || result.Challenges[0] == nil {
			continue
		}
		recommendation := result.Challenges[0]
		return &dto.TeacherRecommendationItem{
			ChallengeID: recommendation.ID,
			Title:       recommendation.Title,
			Category:    recommendation.Category,
			Difficulty:  recommendation.Difficulty,
			Reason:      recommendation.Reason,
		}
	}
	return nil
}
