package queries

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	teachingevidence "ctf-platform/internal/teaching/evidence"
)

type studentReviewQueryRepository interface {
	queryports.TeachingStudentProfileRepository
	queryports.TeachingStudentActivityRepository
}

type StudentReviewQueryService struct {
	users                 queryports.TeachingUserLookupRepository
	repo                  studentReviewQueryRepository
	recommendationService assessmentcontracts.RecommendationProvider
}

var _ StudentReviewService = (*StudentReviewQueryService)(nil)

func NewStudentReviewService(
	users queryports.TeachingUserLookupRepository,
	repo studentReviewQueryRepository,
	recommendationService assessmentcontracts.RecommendationProvider,
) *StudentReviewQueryService {
	return &StudentReviewQueryService{
		users:                 users,
		repo:                  repo,
		recommendationService: recommendationService,
	}
}

func (s *StudentReviewQueryService) GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*TeacherProgressResp, error) {
	student, err := getAccessibleStudent(ctx, s.users, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	totalChallenges, err := s.repo.CountPublishedChallenges(ctx)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	solvedChallenges, err := s.repo.CountSolvedChallenges(ctx, student.ID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	categoryRows, err := s.repo.GetCategoryProgress(ctx, student.ID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	difficultyRows, err := s.repo.GetDifficultyProgress(ctx, student.ID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &TeacherProgressResp{
		TotalChallenges:  int(totalChallenges),
		SolvedChallenges: int(solvedChallenges),
		ByCategory:       toProgressBreakdownMap(categoryRows),
		ByDifficulty:     toProgressBreakdownMap(difficultyRows),
	}, nil
}

func (s *StudentReviewQueryService) GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) (*TeacherRecommendationResp, error) {
	student, err := getAccessibleStudent(ctx, s.users, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	if s.recommendationService == nil {
		return &TeacherRecommendationResp{}, nil
	}

	result, err := s.recommendationService.Recommend(ctx, student.ID, limit)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if result == nil {
		return &TeacherRecommendationResp{}, nil
	}
	resp := teachingQueryMapper.ToTeacherRecommendationRespPtr(result)
	if resp == nil {
		return &TeacherRecommendationResp{}, nil
	}
	resp.WeakDimensions = nonNilSlice(resp.WeakDimensions)
	resp.Challenges = nonNilSlice(resp.Challenges)
	return resp, nil
}

func (s *StudentReviewQueryService) GetStudentTimeline(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit, offset int) (*TimelineResp, error) {
	student, err := getAccessibleStudent(ctx, s.users, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.GetStudentTimeline(ctx, student.ID, limit, offset)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &TimelineResp{Events: nonNilSlice(teachingQueryMapper.ToTimelineEvents(events))}, nil
}

func (s *StudentReviewQueryService) GetStudentEvidence(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *TeacherEvidenceInput) (*TeacherEvidenceResp, error) {
	student, err := getAccessibleStudent(ctx, s.users, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	repoQuery := buildEvidenceQuery(query)
	events, err := s.repo.GetStudentEvidence(ctx, student.ID, repoQuery)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	events = filterEvidenceEvents(events, repoQuery)
	events = paginateEvidenceEvents(events, repoQuery)

	resp := &TeacherEvidenceResp{
		Events: make([]TeacherEvidenceEvent, 0, len(events)),
	}
	for _, event := range events {
		resp.Events = append(resp.Events, TeacherEvidenceEvent{
			Type:        event.Type.String(),
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Detail:      event.Detail,
			Timestamp:   event.Timestamp,
			Meta:        event.Meta,
		})
		resp.Summary.TotalEvents++
		switch event.Type {
		case teachingevidence.EventTypeInstanceProxy:
			resp.Summary.ProxyRequestCount++
		case teachingevidence.EventTypeChallengeSubmission, teachingevidence.EventTypeAWDAttackSubmission:
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

func (s *StudentReviewQueryService) GetStudentAttackSessions(ctx context.Context, requesterID int64, requesterRole string, studentID int64, query *TeacherAttackSessionInput) (*TeacherAttackSessionResp, error) {
	student, err := getAccessibleStudent(ctx, s.users, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	var evidenceInput *TeacherEvidenceInput
	if query != nil {
		evidenceInput = &TeacherEvidenceInput{
			ChallengeID: query.ChallengeID,
			ContestID:   query.ContestID,
			RoundID:     query.RoundID,
		}
	}
	evidenceQuery := buildEvidenceQuery(evidenceInput)
	events, err := s.repo.GetStudentEvidence(ctx, student.ID, evidenceQuery)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
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

	resp := &TeacherAttackSessionResp{
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

func buildEvidenceQuery(query *TeacherEvidenceInput) teachingevidence.Query {
	if query == nil {
		return teachingevidence.Query{}
	}
	return teachingevidence.Query{
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

func filterEvidenceEvents(events []queryports.EvidenceEventRecord, query teachingevidence.Query) []queryports.EvidenceEventRecord {
	if len(events) == 0 {
		return events
	}
	filtered := make([]queryports.EvidenceEventRecord, 0, len(events))
	for _, event := range events {
		if query.ContestID != nil && !int64PtrEqual(event.ContestID, *query.ContestID) {
			continue
		}
		if query.RoundID != nil && !int64PtrEqual(event.RoundID, *query.RoundID) {
			continue
		}
		if query.EventType != "" && event.Type.String() != query.EventType {
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

func paginateEvidenceEvents(events []queryports.EvidenceEventRecord, query teachingevidence.Query) []queryports.EvidenceEventRecord {
	if len(events) == 0 {
		return []queryports.EvidenceEventRecord{}
	}
	offset := query.Offset
	if offset >= len(events) {
		return []queryports.EvidenceEventRecord{}
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

func filterAttackSessionEvents(events []queryports.EvidenceEventRecord, query *TeacherAttackSessionInput) []queryports.EvidenceEventRecord {
	if query == nil {
		return events
	}
	filtered := make([]queryports.EvidenceEventRecord, 0, len(events))
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

func buildAttackSessions(studentID int64, events []queryports.EvidenceEventRecord) []TeacherAttackSession {
	if len(events) == 0 {
		return []TeacherAttackSession{}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	grouped := make(map[string][]queryports.EvidenceEventRecord)
	order := make([]string, 0)
	for _, event := range events {
		key := attackSessionGroupKey(event)
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], event)
	}

	sessions := make([]TeacherAttackSession, 0, len(grouped))
	for _, key := range order {
		chunk := make([]queryports.EvidenceEventRecord, 0)
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

func buildAttackSession(studentID int64, sequence int, events []queryports.EvidenceEventRecord) TeacherAttackSession {
	first := events[0]
	last := events[len(events)-1]
	sessionID := fmt.Sprintf("sess_%d_%d", studentID, sequence)
	session := TeacherAttackSession{
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
		Events:       make([]TeacherAttackEvent, 0, len(events)),
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

func toAttackEvent(studentID int64, sessionID string, index int, event queryports.EvidenceEventRecord) TeacherAttackEvent {
	return TeacherAttackEvent{
		ID:         fmt.Sprintf("%s_evt_%d", sessionID, index+1),
		SessionID:  sessionID,
		Type:       event.Type.String(),
		Stage:      evidenceEventStage(event),
		Source:     evidenceEventSource(event),
		OccurredAt: event.Timestamp,
		Actor: TeacherAttackActor{
			UserID: studentID,
			TeamID: event.TeamID,
		},
		Target: TeacherAttackTarget{
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

func summarizeAttackSessions(sessions []TeacherAttackSession) TeacherAttackSessionSummary {
	summary := TeacherAttackSessionSummary{TotalSessions: len(sessions)}
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

func paginateAttackSessions(sessions []TeacherAttackSession, query *TeacherAttackSessionInput) []TeacherAttackSession {
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
		return []TeacherAttackSession{}
	}
	end := offset + limit
	if end > len(sessions) {
		end = len(sessions)
	}
	return sessions[offset:end]
}

func attackSessionGroupKey(event queryports.EvidenceEventRecord) string {
	if attackEventMode(event) == "awd" {
		return fmt.Sprintf("awd:%s:%s:%s:%s", ptrKey(event.TeamID), ptrKey(event.ContestID), ptrKey(event.ServiceID), ptrKey(event.VictimTeamID))
	}
	return fmt.Sprintf("%s:%d:%s", attackEventMode(event), event.ChallengeID, ptrKey(event.ContestID))
}

func attackEventMode(event queryports.EvidenceEventRecord) string {
	if strings.HasPrefix(event.Type.String(), "awd_") {
		return "awd"
	}
	if event.ContestID != nil {
		return "jeopardy"
	}
	return "practice"
}

func attackEventChallengeID(event queryports.EvidenceEventRecord) *int64 {
	if event.ChallengeID <= 0 {
		return nil
	}
	value := event.ChallengeID
	return &value
}

func deriveAttackSessionResult(events []queryports.EvidenceEventRecord) string {
	hasFailure := false
	hasAction := false
	for _, event := range events {
		switch event.Type {
		case teachingevidence.EventTypeChallengeSubmission, teachingevidence.EventTypeAWDAttackSubmission:
			if evidenceEventSucceeded(event) {
				return "success"
			}
			hasFailure = true
		case teachingevidence.EventTypeInstanceAccess, teachingevidence.EventTypeInstanceProxy, teachingevidence.EventTypeAWDTraffic:
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

func evidenceEventSucceeded(event queryports.EvidenceEventRecord) bool {
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

func evidenceEventStage(event queryports.EvidenceEventRecord) string {
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

func evidenceEventSource(event queryports.EvidenceEventRecord) string {
	if event.Source != "" {
		return event.Source
	}
	switch event.Type {
	case teachingevidence.EventTypeInstanceAccess, teachingevidence.EventTypeInstanceProxy:
		return "audit_logs"
	case teachingevidence.EventTypeChallengeSubmission, teachingevidence.EventTypeManualReview:
		return "submissions"
	case teachingevidence.EventTypeWriteup:
		return "submission_writeups"
	case teachingevidence.EventTypeAWDAttackSubmission:
		return "awd_attack_logs"
	case teachingevidence.EventTypeAWDTraffic:
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

func getAccessibleStudent(
	ctx context.Context,
	repo classAccessRepository,
	requesterID int64,
	requesterRole string,
	studentID int64,
) (*identitycontracts.User, error) {
	student, err := repo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != identitycontracts.RoleStudent {
		return nil, apperror.ErrNotFound
	}

	if err := ensureClassAccess(ctx, repo, requesterID, requesterRole, student.ClassName); err != nil {
		return nil, err
	}
	return student, nil
}

func toProgressBreakdownMap(rows []queryports.ProgressRow) map[string]ProgressBreakdown {
	if len(rows) == 0 {
		return map[string]ProgressBreakdown{}
	}

	result := make(map[string]ProgressBreakdown, len(rows))
	for _, row := range rows {
		result[row.Key] = ProgressBreakdown{
			Total:  row.Total,
			Solved: row.Solved,
		}
	}
	return result
}
