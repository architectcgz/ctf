package commands

import (
	"context"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
	teachingevidence "ctf-platform/internal/teaching/evidence"
)

func (s *ReportService) buildStudentReviewArchiveData(ctx context.Context, studentID int64) (*ReviewArchiveData, error) {
	student, err := s.userRepo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}

	stats, err := s.personalRepo.GetPersonalStats(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	totalChallenges, err := s.reviewArchiveRepo.CountPublishedChallenges(ctx)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	timeline, err := s.reviewArchiveRepo.GetStudentTimeline(ctx, studentID, 200, 0)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	evidence, err := s.reviewArchiveRepo.GetStudentEvidence(ctx, studentID, teachingevidence.Query{})
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	writeups, err := s.reviewArchiveRepo.ListStudentWriteups(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	manualReviews, err := s.reviewArchiveRepo.ListStudentManualReviews(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	var skillProfile []*assessmentcontracts.SkillDimension
	if s.assessmentService != nil {
		skillProfileResp, skillErr := s.assessmentService.GetSkillProfile(ctx, studentID)
		if skillErr != nil {
			return nil, apperror.ErrInternal.WithCause(skillErr)
		}
		skillProfile = skillProfileResp.Dimensions
	}

	summary := buildReviewArchiveSummary(int(totalChallenges), stats, timeline, evidence, writeups, manualReviews)

	return &ReviewArchiveData{
		GeneratedAt: time.Now().UTC(),
		Student: ReviewArchiveStudent{
			ID:        student.ID,
			Username:  student.Username,
			Name:      student.Name,
			ClassName: student.ClassName,
		},
		Summary:             summary,
		SkillProfile:        skillProfile,
		Timeline:            timeline,
		Evidence:            evidence,
		Writeups:            writeups,
		ManualReviews:       manualReviews,
		TeacherObservations: buildReviewArchiveObservations(summary, skillProfile, timeline, evidence, writeups, manualReviews),
	}, nil
}

func buildReviewArchiveSummary(
	totalChallenges int,
	stats *assessmentdomain.PersonalReportStats,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) assessmentdomain.ReviewArchiveSummary {
	summary := assessmentdomain.ReviewArchiveSummary{
		TotalChallenges:        totalChallenges,
		TimelineEventCount:     len(timeline),
		EvidenceEventCount:     len(evidence),
		WriteupCount:           len(writeups),
		ManualReviewCount:      len(manualReviews),
		CorrectSubmissionCount: countCorrectSubmissions(timeline, evidence),
		LastActivityAt:         latestReviewArchiveActivity(timeline, evidence, writeups, manualReviews),
	}
	if stats != nil {
		summary.TotalSolved = stats.TotalSolved
		summary.TotalScore = stats.TotalScore
		summary.Rank = stats.Rank
		summary.TotalAttempts = stats.TotalAttempts
	}
	return summary
}

func countCorrectSubmissions(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) int {
	if stats, ok := reviewArchiveSubmissionStatsFromEvidence(evidence); ok {
		if !stats.HasChallengeEvidence {
			challengeSuccessCount := countCorrectTimelineChallengeSubmissions(timeline)
			stats.ChallengeSuccessCount += challengeSuccessCount
			stats.SuccessCount += challengeSuccessCount
		}
		if !stats.HasAWDEvidence {
			awdSuccessCount := countCorrectTimelineAWDSubmissions(timeline)
			stats.AWDSuccessCount += awdSuccessCount
			stats.SuccessCount += awdSuccessCount
		}
		return stats.SuccessCount
	}
	return countCorrectTimelineChallengeSubmissions(timeline) + countCorrectTimelineAWDSubmissions(timeline)
}

func latestReviewArchiveActivity(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) *time.Time {
	var latest *time.Time
	record := func(candidate *time.Time) {
		if candidate == nil || candidate.IsZero() {
			return
		}
		if latest == nil || candidate.After(*latest) {
			copyValue := *candidate
			latest = &copyValue
		}
	}

	for _, item := range timeline {
		record(&item.Timestamp)
	}
	for _, item := range evidence {
		if !includeEvidenceInPersonalActivity(item) {
			continue
		}
		record(&item.Timestamp)
	}
	for _, item := range writeups {
		if item.PublishedAt != nil {
			record(item.PublishedAt)
			continue
		}
		record(&item.UpdatedAt)
	}
	for _, item := range manualReviews {
		record(&item.SubmittedAt)
	}
	return latest
}

func buildReviewArchiveObservations(
	summary assessmentdomain.ReviewArchiveSummary,
	skillProfile []*assessmentcontracts.SkillDimension,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) assessmentdomain.ReviewArchiveTeacherObservations {
	snapshot := buildReviewArchiveTeachingFactSnapshot(summary, skillProfile, timeline, evidence, writeups, manualReviews)
	evaluation := teachingadvice.EvaluateStudent(snapshot)
	adviceItems := teachingadvice.BuildReviewArchiveObservations(snapshot, evaluation)

	items := make([]assessmentdomain.ReviewArchiveObservation, 0, len(adviceItems))
	for _, item := range adviceItems {
		items = append(items, assessmentdomain.ReviewArchiveObservation{
			Code:      item.Code,
			Label:     item.Label,
			Severity:  string(item.Severity),
			Dimension: item.Dimension,
			Summary:   item.Summary,
			Evidence:  item.Evidence,
			Action:    item.Action,
		})
	}
	return assessmentdomain.ReviewArchiveTeacherObservations{Items: items}
}

func buildReviewArchiveTeachingFactSnapshot(
	summary assessmentdomain.ReviewArchiveSummary,
	skillProfile []*assessmentcontracts.SkillDimension,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) teachingadvice.StudentFactSnapshot {
	recentEventCount, activeDays := recentReviewArchiveActivityStats(time.Now().UTC(), timeline, evidence, writeups, manualReviews)
	submissionStats := buildReviewArchiveSubmissionStats(summary, timeline, evidence)
	snapshot := teachingadvice.StudentFactSnapshot{
		ActiveDays7d:           activeDays,
		RecentEventCount7d:     recentEventCount,
		LastActivityAt:         summary.LastActivityAt,
		CorrectSubmissionCount: submissionStats.SuccessCount,
		WrongSubmissionCount:   submissionStats.FailureCount,
		ChallengeSuccessCount:  submissionStats.ChallengeSuccessCount,
		SubmissionSuccessCount: submissionStats.SuccessCount,
		SubmissionFailureCount: submissionStats.FailureCount,
		WriteupCount:           summary.WriteupCount,
		ApprovedReviewCount:    countApprovedManualReviews(manualReviews),
		Dimensions:             make([]teachingadvice.DimensionFact, 0, len(taxonomy.AllDimensions)),
	}

	factMap := make(map[string]*teachingadvice.DimensionFact, len(taxonomy.AllDimensions))
	for _, dimension := range taxonomy.AllDimensions {
		dimensionCopy := dimension
		factMap[dimension] = &teachingadvice.DimensionFact{Dimension: dimensionCopy}
	}

	snapshot.MaxWrongStreak = submissionStats.MaxWrongStreak
	snapshot.HandsOnEventCount = countReviewArchiveAWDHandsOnEvidence(timeline, evidence)
	snapshot.AWDAttemptCount = countReviewArchiveAWDAttempts(timeline, evidence)
	snapshot.AWDSuccessCount = submissionStats.AWDSuccessCount

	for _, dimension := range skillProfile {
		if dimension == nil {
			continue
		}
		fact := ensureReviewArchiveDimensionFact(factMap, dimension.Dimension)
		if fact == nil {
			continue
		}
		fact.ProfileScore = dimension.Score
	}

	for _, item := range evidence {
		fact := ensureReviewArchiveDimensionFact(factMap, item.Category)
		if fact == nil {
			continue
		}
		switch item.Type {
		case teachingevidence.EventTypeChallengeSubmission, teachingevidence.EventTypeAWDAttackSubmission:
			if item.Type == teachingevidence.EventTypeAWDAttackSubmission && !isStudentScopedAWDAttackEvidence(item) {
				continue
			}
			fact.AttemptCount++
			if success, tracked := extractEvidenceSubmissionResult(item); tracked && success {
				fact.SuccessCount++
			}
			fact.EvidenceCount++
		case teachingevidence.EventTypeInstanceAccess, teachingevidence.EventTypeInstanceProxy, teachingevidence.EventTypeAWDTraffic:
			fact.EvidenceCount++
		}
	}

	for _, item := range writeups {
		fact := ensureReviewArchiveDimensionFact(factMap, item.Category)
		if fact == nil {
			continue
		}
		fact.EvidenceCount++
	}

	for _, item := range manualReviews {
		if item.ReviewStatus != "approved" {
			continue
		}
		fact := ensureReviewArchiveDimensionFact(factMap, item.Category)
		if fact == nil {
			continue
		}
		fact.EvidenceCount++
	}

	for _, dimension := range taxonomy.AllDimensions {
		fact := ensureReviewArchiveDimensionFact(factMap, dimension)
		if fact == nil {
			continue
		}
		snapshot.Dimensions = append(snapshot.Dimensions, *fact)
	}

	return snapshot
}

func recentReviewArchiveActivityStats(
	referenceTime time.Time,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) (int, int) {
	if referenceTime.IsZero() {
		referenceTime = time.Now().UTC()
	}
	cutoff := referenceTime.AddDate(0, 0, -7)
	activeDays := make(map[string]struct{})
	recentEventCount := 0

	record := func(timestamp time.Time) {
		if timestamp.IsZero() || timestamp.Before(cutoff) {
			return
		}
		recentEventCount++
		activeDays[timestamp.UTC().Format("2006-01-02")] = struct{}{}
	}

	if len(evidence) > 0 {
		for _, item := range evidence {
			if !includeEvidenceInPersonalActivity(item) {
				continue
			}
			record(item.Timestamp)
		}
	} else {
		for _, item := range timeline {
			record(item.Timestamp)
		}
	}

	for _, item := range writeups {
		if item.PublishedAt != nil {
			record(item.PublishedAt.UTC())
			continue
		}
		record(item.UpdatedAt)
	}

	for _, item := range manualReviews {
		record(item.SubmittedAt)
	}

	return recentEventCount, len(activeDays)
}

func ensureReviewArchiveDimensionFact(
	facts map[string]*teachingadvice.DimensionFact,
	dimension string,
) *teachingadvice.DimensionFact {
	normalized := strings.ToLower(strings.TrimSpace(dimension))
	if normalized == "" {
		return nil
	}
	if fact, ok := facts[normalized]; ok {
		return fact
	}
	fact := &teachingadvice.DimensionFact{Dimension: normalized}
	facts[normalized] = fact
	return fact
}

func hasSubmittedWriteup(writeups []assessmentdomain.ReviewArchiveWriteupItem) bool {
	for _, item := range writeups {
		if item.SubmissionStatus == "published" || item.SubmissionStatus == "submitted" {
			return true
		}
	}
	return false
}

func countApprovedManualReviews(items []assessmentdomain.ReviewArchiveManualReviewItem) int {
	count := 0
	for _, item := range items {
		if item.ReviewStatus == "approved" {
			count++
		}
	}
	return count
}

func hasApprovedManualReview(items []assessmentdomain.ReviewArchiveManualReviewItem) bool {
	return countApprovedManualReviews(items) > 0
}

func hasRepeatedWrongSubmissions(evidence []assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	streak := 0
	for _, item := range evidence {
		isCorrect, tracked := extractEvidenceSubmissionResult(item)
		if !tracked {
			continue
		}
		if isCorrect {
			streak = 0
			continue
		}
		streak++
		if streak >= 2 {
			return true
		}
	}
	return false
}

func hasHandsOnExploit(evidence []assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	for _, item := range evidence {
		if item.Type == teachingevidence.EventTypeInstanceAccess ||
			item.Type == teachingevidence.EventTypeInstanceProxy ||
			item.Type == teachingevidence.EventTypeAWDTraffic {
			return true
		}
		if item.Type == teachingevidence.EventTypeAWDAttackSubmission && isStudentScopedAWDAttackEvidence(item) {
			return true
		}
	}
	return false
}

func isCorrectTimelineSubmission(item assessmentdomain.ReviewArchiveTimelineEvent) bool {
	if item.IsCorrect == nil || !*item.IsCorrect {
		return false
	}
	return item.Type == "flag_submit" || item.Type == "awd_attack_submit"
}

func isCorrectEvidenceSubmission(item assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	isCorrect, tracked := extractEvidenceSubmissionResult(item)
	return tracked && isCorrect
}

func extractEvidenceSubmissionResult(item assessmentdomain.ReviewArchiveEvidenceEvent) (bool, bool) {
	if item.Meta == nil {
		return false, false
	}

	switch item.Type {
	case teachingevidence.EventTypeChallengeSubmission:
		isCorrect, ok := item.Meta["is_correct"].(bool)
		return isCorrect, ok
	case teachingevidence.EventTypeAWDAttackSubmission:
		if !isStudentScopedAWDAttackEvidence(item) {
			return false, false
		}
		isCorrect, ok := item.Meta["is_success"].(bool)
		return isCorrect, ok
	default:
		return false, false
	}
}

type reviewArchiveSubmissionStats struct {
	ChallengeSuccessCount int
	SuccessCount          int
	FailureCount          int
	AWDSuccessCount       int
	MaxWrongStreak        int
	HasChallengeEvidence  bool
	HasAWDEvidence        bool
}

func buildReviewArchiveSubmissionStats(
	summary assessmentdomain.ReviewArchiveSummary,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) reviewArchiveSubmissionStats {
	if stats, ok := reviewArchiveSubmissionStatsFromEvidence(evidence); ok {
		if !stats.HasChallengeEvidence {
			challengeSuccessCount := countCorrectTimelineChallengeSubmissions(timeline)
			stats.ChallengeSuccessCount += challengeSuccessCount
			stats.SuccessCount += challengeSuccessCount
			stats.FailureCount += max(summary.TotalAttempts-stats.ChallengeSuccessCount, 0)
		}
		if !stats.HasAWDEvidence {
			awdSuccessCount := countCorrectTimelineAWDSubmissions(timeline)
			stats.AWDSuccessCount += awdSuccessCount
			stats.SuccessCount += awdSuccessCount
		}
		return stats
	}

	stats := reviewArchiveSubmissionStats{}
	stats.ChallengeSuccessCount = countCorrectTimelineChallengeSubmissions(timeline)
	stats.AWDSuccessCount = countCorrectTimelineAWDSubmissions(timeline)
	stats.SuccessCount = stats.ChallengeSuccessCount + stats.AWDSuccessCount
	stats.FailureCount = max(summary.TotalAttempts-stats.ChallengeSuccessCount, 0)
	return stats
}

func reviewArchiveSubmissionStatsFromEvidence(
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) (reviewArchiveSubmissionStats, bool) {
	type trackedEvent struct {
		timestamp time.Time
		success   bool
	}

	stats := reviewArchiveSubmissionStats{}
	trackedEvents := make([]trackedEvent, 0, len(evidence))
	trackedCount := 0

	for _, item := range evidence {
		isCorrect, tracked := extractEvidenceSubmissionResult(item)
		if !tracked {
			continue
		}
		trackedCount++
		trackedEvents = append(trackedEvents, trackedEvent{timestamp: item.Timestamp, success: isCorrect})
		if item.Type == teachingevidence.EventTypeChallengeSubmission {
			stats.HasChallengeEvidence = true
		}
		if item.Type == teachingevidence.EventTypeAWDAttackSubmission {
			stats.HasAWDEvidence = true
		}
		if isCorrect {
			stats.SuccessCount++
			if item.Type == teachingevidence.EventTypeChallengeSubmission {
				stats.ChallengeSuccessCount++
			}
			if item.Type == teachingevidence.EventTypeAWDAttackSubmission {
				stats.AWDSuccessCount++
			}
			continue
		}
		stats.FailureCount++
	}

	if trackedCount == 0 {
		return reviewArchiveSubmissionStats{}, false
	}

	sort.Slice(trackedEvents, func(i, j int) bool {
		return trackedEvents[i].timestamp.Before(trackedEvents[j].timestamp)
	})

	currentWrongStreak := 0
	for _, event := range trackedEvents {
		if event.success {
			currentWrongStreak = 0
			continue
		}
		currentWrongStreak++
		if currentWrongStreak > stats.MaxWrongStreak {
			stats.MaxWrongStreak = currentWrongStreak
		}
	}

	return stats, true
}

func countCorrectTimelineChallengeSubmissions(timeline []assessmentdomain.ReviewArchiveTimelineEvent) int {
	count := 0
	for _, item := range timeline {
		if item.Type == "flag_submit" && item.IsCorrect != nil && *item.IsCorrect {
			count++
		}
	}
	return count
}

func countReviewArchiveAWDHandsOnEvidence(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) int {
	handsOnCount := 0
	for _, item := range evidence {
		switch item.Type {
		case teachingevidence.EventTypeAWDTraffic:
			handsOnCount++
		case teachingevidence.EventTypeAWDAttackSubmission:
			if isStudentScopedAWDAttackEvidence(item) {
				handsOnCount++
			}
		}
	}
	if handsOnCount > 0 {
		return handsOnCount
	}
	for _, item := range timeline {
		if item.Type == "awd_attack_submit" {
			handsOnCount++
		}
	}
	return handsOnCount
}

func countReviewArchiveAWDAttempts(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) int {
	attemptCount := 0
	for _, item := range evidence {
		if item.Type == teachingevidence.EventTypeAWDAttackSubmission && isStudentScopedAWDAttackEvidence(item) {
			attemptCount++
		}
	}
	if attemptCount > 0 {
		return attemptCount
	}
	for _, item := range timeline {
		if item.Type == "awd_attack_submit" {
			attemptCount++
		}
	}
	return attemptCount
}

func includeEvidenceInPersonalActivity(item assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	switch item.Type {
	case teachingevidence.EventTypeAWDTraffic:
		return false
	case teachingevidence.EventTypeAWDAttackSubmission:
		return isStudentScopedAWDAttackEvidence(item)
	default:
		return true
	}
}

func isStudentScopedAWDAttackEvidence(item assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	if item.Type != teachingevidence.EventTypeAWDAttackSubmission {
		return false
	}
	if item.Meta == nil {
		return true
	}
	scope, ok := item.Meta["scope"].(string)
	if !ok {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(scope), "student")
}

func countCorrectTimelineAWDSubmissions(timeline []assessmentdomain.ReviewArchiveTimelineEvent) int {
	count := 0
	for _, item := range timeline {
		if item.Type == "awd_attack_submit" && item.IsCorrect != nil && *item.IsCorrect {
			count++
		}
	}
	return count
}
