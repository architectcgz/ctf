package advice

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/model"
)

type Severity string

const (
	SeverityGood      Severity = "good"
	SeverityAttention Severity = "attention"
	SeverityWarning   Severity = "warning"
	SeverityDanger    Severity = "danger"
)

type DifficultyBand string

const (
	DifficultyBandBeginner DifficultyBand = "beginner"
	DifficultyBandEasy     DifficultyBand = "easy"
	DifficultyBandMedium   DifficultyBand = "medium"
	DifficultyBandHard     DifficultyBand = "hard"
	DifficultyBandInsane   DifficultyBand = "insane"
)

type StudentFactSnapshot struct {
	UserID                 int64
	Username               string
	Name                   *string
	ActiveDays7d           int
	RecentEventCount7d     int
	LastActivityAt         *time.Time
	CorrectSubmissionCount int
	WrongSubmissionCount   int
	ChallengeSuccessCount  int
	SubmissionSuccessCount int
	SubmissionFailureCount int
	MaxWrongStreak         int
	WriteupCount           int
	ApprovedReviewCount    int
	HandsOnEventCount      int
	AWDSuccessCount        int
	Dimensions             []DimensionFact
}

type DimensionFact struct {
	Dimension              string
	ProfileScore           float64
	AttemptCount           int
	SuccessCount           int
	EvidenceCount          int
	SolvedDifficultyCounts map[string]int
}

type DimensionAdvice struct {
	Dimension     string
	Confidence    float64
	IsWeak        bool
	Severity      Severity
	ProfileScore  float64
	AttemptCount  int
	SuccessCount  int
	EvidenceCount int
	ReasonCodes   []string
	Summary       string
	Evidence      string
}

type StudentEvaluation struct {
	Severity                  Severity
	Dimensions                []DimensionAdvice
	WeakDimensions            []DimensionAdvice
	RecommendationTargets     []DimensionAdvice
	RecommendedDifficultyBand DifficultyBand
}

type ReviewArchiveObservation struct {
	Code      string
	Label     string
	Severity  Severity
	Dimension *string
	Summary   string
	Evidence  string
	Action    string
}

type ClassSummarySnapshot struct {
	ClassName        string
	StudentCount     int
	ActiveRate       float64
	RecentEventCount int64
}

type ClassTrendSnapshot struct {
	EventDelta int64
	SolveDelta int64
}

type ClassReviewItem struct {
	Code                    string
	Severity                Severity
	Summary                 string
	Evidence                string
	Action                  string
	ReasonCodes             []string
	Dimension               string
	StudentIDs              []int64
	RecommendationStudentID *int64
}

type ChallengeCandidate struct {
	ID         int64
	Title      string
	Category   string
	Dimension  string
	Difficulty string
	Points     int
}

type RecommendationReason struct {
	Dimension      string
	DifficultyBand DifficultyBand
	Severity       Severity
	ReasonCodes    []string
	Summary        string
	Evidence       string
}

type RecommendationPlan struct {
	WeakDimensions []DimensionAdvice
	Targets        []DimensionAdvice
	Reasons        []RecommendationReason
	DifficultyBand DifficultyBand
}

func EvaluateStudent(snapshot StudentFactSnapshot) StudentEvaluation {
	dimensions := make([]DimensionAdvice, 0, len(snapshot.Dimensions))
	for _, fact := range snapshot.Dimensions {
		advice := evaluateDimension(fact)
		if advice.Dimension == "" {
			continue
		}
		dimensions = append(dimensions, advice)
	}

	sort.Slice(dimensions, func(i, j int) bool {
		if severityRank(dimensions[i].Severity) != severityRank(dimensions[j].Severity) {
			return severityRank(dimensions[i].Severity) > severityRank(dimensions[j].Severity)
		}
		if dimensions[i].Confidence != dimensions[j].Confidence {
			return dimensions[i].Confidence > dimensions[j].Confidence
		}
		if dimensions[i].ProfileScore != dimensions[j].ProfileScore {
			return dimensions[i].ProfileScore < dimensions[j].ProfileScore
		}
		return dimensions[i].Dimension < dimensions[j].Dimension
	})

	weakDimensions := make([]DimensionAdvice, 0, len(dimensions))
	recommendationTargets := make([]DimensionAdvice, 0, 2)
	for _, item := range dimensions {
		if item.IsWeak {
			weakDimensions = append(weakDimensions, item)
			recommendationTargets = append(recommendationTargets, item)
			continue
		}
		if len(recommendationTargets) == 0 && item.EvidenceCount > 0 && item.ProfileScore < 0.6 {
			recommendationTargets = append(recommendationTargets, item)
		}
	}

	if len(recommendationTargets) == 0 {
		for _, item := range dimensions {
			if item.Severity == SeverityGood {
				continue
			}
			if item.EvidenceCount > 0 || item.AttemptCount > 0 {
				recommendationTargets = append(recommendationTargets, item)
				break
			}
		}
	}

	progressionBand := DifficultyBand("")
	if len(recommendationTargets) == 0 {
		if target, band := selectProgressionTarget(snapshot.Dimensions, dimensions); target.Dimension != "" {
			recommendationTargets = append(recommendationTargets, target)
			progressionBand = band
		}
	}

	difficultyBand := DifficultyBandMedium
	switch {
	case len(weakDimensions) > 0 && weakDimensions[0].Severity == SeverityDanger:
		difficultyBand = DifficultyBandBeginner
	case snapshot.MaxWrongStreak >= 4:
		difficultyBand = DifficultyBandBeginner
	case len(weakDimensions) > 0:
		difficultyBand = DifficultyBandEasy
	case snapshot.MaxWrongStreak >= 2:
		difficultyBand = DifficultyBandEasy
	case len(recommendationTargets) > 0:
		difficultyBand = DifficultyBandEasy
	}
	if progressionBand != "" {
		difficultyBand = progressionBand
	}

	severity := SeverityGood
	if len(weakDimensions) > 0 {
		severity = maxSeverity(severity, weakDimensions[0].Severity)
	}
	if submissionSeverity, ok := submissionStabilitySeverity(snapshot); ok {
		severity = maxSeverity(severity, submissionSeverity)
	}
	if challengeSuccessCount(snapshot) > 0 && snapshot.WriteupCount+snapshot.ApprovedReviewCount == 0 {
		severity = maxSeverity(severity, SeverityAttention)
	}
	if activitySeverity, ok := lowActivitySeverity(snapshot); ok {
		severity = maxSeverity(severity, activitySeverity)
	}

	return StudentEvaluation{
		Severity:                  severity,
		Dimensions:                dimensions,
		WeakDimensions:            weakDimensions,
		RecommendationTargets:     recommendationTargets,
		RecommendedDifficultyBand: difficultyBand,
	}
}

func BuildReviewArchiveObservations(snapshot StudentFactSnapshot, evaluation StudentEvaluation) []ReviewArchiveObservation {
	items := make([]ReviewArchiveObservation, 0, 5)
	challengeSuccessCount := challengeSuccessCount(snapshot)
	submissionSuccessCount := submissionSuccessCount(snapshot)
	submissionFailureCount := submissionFailureCount(snapshot)

	if challengeSuccessCount > 0 {
		outputCount := snapshot.WriteupCount + snapshot.ApprovedReviewCount
		observation := ReviewArchiveObservation{
			Code:     "training_closure",
			Label:    "训练闭环",
			Severity: SeverityAttention,
			Summary:  "已经拿到训练结果，但复盘输出还没有稳定跟上。",
			Evidence: buildTrainingClosureEvidence(challengeSuccessCount, submissionSuccessCount, snapshot.AWDSuccessCount, snapshot.WriteupCount, snapshot.ApprovedReviewCount),
			Action:   "补 1 份复盘材料或课堂讲解记录，把成功经验沉淀下来。",
		}
		if outputCount >= challengeSuccessCount {
			observation.Severity = SeverityGood
			observation.Summary = "已经形成从解题到复盘输出的训练闭环。"
			observation.Action = "继续保持输出质量，把高质量复盘沉淀成可复用材料。"
		} else if outputCount > 0 {
			observation.Summary = "已经开始补复盘输出，但还没有覆盖到多数成功样本。"
			observation.Action = "优先把最近一次成功过程写透，再逐步补齐其他关键样本。"
		}
		items = append(items, observation)
	}

	if severity, ok := submissionStabilitySeverity(snapshot); ok {
		summary := "连续失败事件偏多，当前还没有形成稳定命中。"
		if submissionSuccessCount > 0 {
			summary = "已经有成功结果，但重复失败仍在抬高试错成本。"
		}
		items = append(items, ReviewArchiveObservation{
			Code:     "submission_stability",
			Label:    "提交稳定性",
			Severity: severity,
			Summary:  summary,
			Evidence: fmt.Sprintf("成功事件 %d 次，失败事件 %d 次，最长连续失败 %d 次。", submissionSuccessCount, submissionFailureCount, snapshot.MaxWrongStreak),
			Action:   "先回看关键一步的利用思路，再继续提交，避免把时间消耗在重复试错上。",
		})
	} else if submissionSuccessCount > 0 {
		summary := "提交节奏整体稳定，归档里没有明显的失败积累。"
		if submissionFailureCount > 0 {
			summary = "有少量试错，但整体还能稳定收敛到结果。"
		}
		items = append(items, ReviewArchiveObservation{
			Code:     "submission_stability",
			Label:    "提交稳定性",
			Severity: SeverityGood,
			Summary:  summary,
			Evidence: fmt.Sprintf("成功事件 %d 次，失败事件 %d 次。", submissionSuccessCount, submissionFailureCount),
			Action:   "继续保持先验证思路、再提交结果的节奏。",
		})
	}

	if observation, ok := buildLowActivityObservation(snapshot); ok {
		items = append(items, observation)
	}

	if snapshot.HandsOnEventCount+snapshot.AWDSuccessCount > 0 {
		severity := SeverityGood
		summary := "已经开始留下实操过程证据，后续要继续把过程收束成结果。"
		switch {
		case snapshot.HandsOnEventCount > 0 && snapshot.AWDSuccessCount > 0:
			summary = "实操过程证据比较完整，而且已经拿到 AWD 实战结果。"
		case snapshot.HandsOnEventCount == 0 && snapshot.AWDSuccessCount > 0:
			summary = "已有 AWD 实战结果，说明能够把技能迁移到攻防场景。"
		case snapshot.HandsOnEventCount >= 3:
			summary = "实操交互证据比较充分，训练过程可复盘。"
		}
		if submissionSuccessCount == 0 && snapshot.AWDSuccessCount == 0 {
			severity = SeverityAttention
		}
		items = append(items, ReviewArchiveObservation{
			Code:     "hands_on_depth",
			Label:    "实操深度",
			Severity: severity,
			Summary:  summary,
			Evidence: fmt.Sprintf("实例/代理交互 %d 次，AWD 成功 %d 次。", snapshot.HandsOnEventCount, snapshot.AWDSuccessCount),
			Action:   "保留关键操作证据，后续复盘时优先回放这类高价值步骤。",
		})
	}

	if len(evaluation.WeakDimensions) > 0 {
		top := evaluation.WeakDimensions[0]
		dimension := top.Dimension
		items = append(items, ReviewArchiveObservation{
			Code:      "dimension_focus",
			Label:     "维度聚焦",
			Severity:  top.Severity,
			Dimension: &dimension,
			Summary:   buildWeakDimensionObservationSummary(top),
			Evidence:  top.Evidence,
			Action:    fmt.Sprintf("接下来优先补 %s 维度的 %s 难度题。", dimensionLabel(top.Dimension), evaluation.RecommendedDifficultyBand),
		})
	} else if len(evaluation.RecommendationTargets) > 0 {
		top := evaluation.RecommendationTargets[0]
		if !containsReasonCode(top.ReasonCodes, "progression_ready") {
			dimension := top.Dimension
			items = append(items, ReviewArchiveObservation{
				Code:      "dimension_focus",
				Label:     "维度聚焦",
				Severity:  SeverityAttention,
				Dimension: &dimension,
				Summary:   buildCoverageGapObservationSummary(top),
				Evidence:  top.Evidence,
				Action:    fmt.Sprintf("先补 1 道 %s 维度的 %s 难度题，把训练样本补齐。", dimensionLabel(top.Dimension), evaluation.RecommendedDifficultyBand),
			})
		}
	}

	if snapshot.AWDSuccessCount > 0 {
		items = append(items, ReviewArchiveObservation{
			Code:     "awd_participation",
			Label:    "AWD 实战参与",
			Severity: SeverityGood,
			Summary:  "已经在 AWD 场景拿到有效结果，具备一定的实战迁移能力。",
			Evidence: fmt.Sprintf("AWD 成功攻击 %d 次。", snapshot.AWDSuccessCount),
			Action:   "后续可以继续用更完整的攻击链复盘，巩固迁移能力。",
		})
	}

	return items
}

func BuildClassReview(
	className string,
	summary ClassSummarySnapshot,
	trend *ClassTrendSnapshot,
	snapshots []StudentFactSnapshot,
	evaluations map[int64]StudentEvaluation,
) []ClassReviewItem {
	items := make([]ClassReviewItem, 0, 5)
	if len(snapshots) == 0 {
		return items
	}

	lowActivityStudents := make([]StudentFactSnapshot, 0)
	closureGapStudents := make([]StudentFactSnapshot, 0)
	retryRiskStudents := make([]StudentFactSnapshot, 0)
	weakDimensionCounts := make(map[string]int)
	weakDimensionStudents := make(map[string][]StudentFactSnapshot)
	weakDimensionSeverities := make(map[string]Severity)

	for _, snapshot := range snapshots {
		if isLowActivity(snapshot) {
			lowActivityStudents = append(lowActivityStudents, snapshot)
		}
		if snapshot.CorrectSubmissionCount > 0 && snapshot.WriteupCount+snapshot.ApprovedReviewCount == 0 {
			closureGapStudents = append(closureGapStudents, snapshot)
		}
		failureCount := submissionFailureCount(snapshot)
		successCount := submissionSuccessCount(snapshot)
		if snapshot.MaxWrongStreak >= 3 || (failureCount >= 5 && failureCount > successCount*2) {
			retryRiskStudents = append(retryRiskStudents, snapshot)
		}

		evaluation := evaluations[snapshot.UserID]
		for _, dimension := range evaluation.WeakDimensions {
			weakDimensionCounts[dimension.Dimension]++
			weakDimensionStudents[dimension.Dimension] = append(weakDimensionStudents[dimension.Dimension], snapshot)
			weakDimensionSeverities[dimension.Dimension] = maxSeverity(weakDimensionSeverities[dimension.Dimension], dimension.Severity)
		}
	}

	sortStudentsByRisk(lowActivityStudents)
	sortStudentsByRisk(closureGapStudents)
	sortStudentsByRisk(retryRiskStudents)

	activitySeverity := SeverityGood
	activitySummary := fmt.Sprintf("%s 最近一周的训练节奏整体稳定。", className)
	activityAction := "继续保持当前训练节奏，优先把注意力放在明确薄弱项和复盘闭环上。"
	lowActivityRatio := float64(len(lowActivityStudents)) / float64(len(snapshots))
	if summary.ActiveRate < 50 || lowActivityRatio >= 0.5 {
		activitySeverity = SeverityDanger
		activitySummary = fmt.Sprintf("%s 最近一周的训练活跃度明显下滑。", className)
		activityAction = "先联系低活跃学生确认卡点，再安排一次短时补练把班级节奏拉回来。"
	} else if summary.ActiveRate < 75 || lowActivityRatio >= 0.25 {
		activitySeverity = SeverityWarning
		activitySummary = fmt.Sprintf("%s 最近一周的训练节奏开始变松。", className)
		activityAction = "本周优先跟进低活跃学生，避免节奏进一步松散。"
	} else if len(lowActivityStudents) > 0 {
		activitySeverity = SeverityAttention
		activitySummary = fmt.Sprintf("%s 整体训练还在推进，但有少数学生近期节奏偏慢。", className)
		activityAction = "点名跟进名单中的学生，优先确认是否存在卡题或掉队风险。"
	}
	items = append(items, ClassReviewItem{
		Code:        "activity_risk",
		Severity:    activitySeverity,
		Summary:     activitySummary,
		Evidence:    fmt.Sprintf("近 7 天活跃率 %.0f%%，低活跃学生 %d/%d，训练事件 %d 次。", summary.ActiveRate, len(lowActivityStudents), len(snapshots), summary.RecentEventCount),
		Action:      activityAction,
		ReasonCodes: []string{"activity_rate", "recent_event_count"},
		StudentIDs:  studentIDs(lowActivityStudents, 3),
	})

	if dimension, students := selectTopWeakDimension(weakDimensionCounts, weakDimensionStudents, len(snapshots)); dimension != "" {
		items = append(items, ClassReviewItem{
			Code:                    "weak_dimension_cluster",
			Severity:                weakDimensionSeverities[dimension],
			Summary:                 fmt.Sprintf("%s 维度是当前最集中的高置信度薄弱项。", dimensionLabel(dimension)),
			Evidence:                fmt.Sprintf("共有 %d 名学生在该维度同时表现为得分偏低且已有足够训练证据。", len(students)),
			Action:                  fmt.Sprintf("本周统一布置 1 到 2 道 %s 维度的 %s 难度题，先补基础命中率。", dimensionLabel(dimension), classDifficultyBand(students, evaluations, dimension)),
			ReasonCodes:             []string{"weak_dimension_cluster", "evidence_sufficient"},
			Dimension:               dimension,
			StudentIDs:              studentIDs(students, 3),
			RecommendationStudentID: firstStudentID(students),
		})
	}

	if len(closureGapStudents) > 0 {
		items = append(items, ClassReviewItem{
			Code:        "training_closure_gap",
			Severity:    SeverityWarning,
			Summary:     "部分学生已经能解题，但训练闭环还没有形成。",
			Evidence:    fmt.Sprintf("共有 %d 名学生出现“有正确提交但缺少 writeup 或通过评阅记录”的情况。", len(closureGapStudents)),
			Action:      "把复盘材料或课堂讲解记录列为本周交付物，优先跟进名单中的学生。",
			ReasonCodes: []string{"closure_gap", "missing_review_output"},
			StudentIDs:  studentIDs(closureGapStudents, 3),
		})
	}

	if len(retryRiskStudents) > 0 {
		items = append(items, ClassReviewItem{
			Code:        "retry_cost_high",
			Severity:    SeverityWarning,
			Summary:     "部分学生的试错成本已经偏高，需要先收紧提交流程。",
			Evidence:    fmt.Sprintf("共有 %d 名学生出现连续错误提交 >= 3 次或错误提交显著高于正确提交。", len(retryRiskStudents)),
			Action:      "先回放利用链路，再要求学生按“验证思路 -> 再提交”的节奏继续训练。",
			ReasonCodes: []string{"wrong_streak", "retry_cost"},
			StudentIDs:  studentIDs(retryRiskStudents, 3),
		})
	}

	if trend != nil {
		severity := SeverityGood
		summaryText := "最近一周训练投入还在向前推进。"
		action := "继续保持当前节奏，把教师介入重点放在薄弱维度和闭环补强上。"
		if trend.EventDelta < 0 || trend.SolveDelta < 0 {
			severity = SeverityWarning
			summaryText = "最近一周训练走势开始回落，需要关注投入是否继续下滑。"
			action = "先稳定训练频次，再安排集中补练，避免班级节奏继续走低。"
		}
		items = append(items, ClassReviewItem{
			Code:        "trend_watch",
			Severity:    severity,
			Summary:     summaryText,
			Evidence:    fmt.Sprintf("最近一周训练事件变化 %d，成功解题变化 %d。", trend.EventDelta, trend.SolveDelta),
			Action:      action,
			ReasonCodes: []string{"trend_event_delta", "trend_solve_delta"},
		})
	}

	return items
}

func BuildRecommendationPlan(snapshot StudentFactSnapshot, evaluation StudentEvaluation, challenges []ChallengeCandidate) RecommendationPlan {
	reasons := make([]RecommendationReason, 0, len(challenges))
	for _, challenge := range challenges {
		target := pickRecommendationTarget(recommendationTargetDimension(challenge), evaluation)
		if target.Dimension == "" {
			continue
		}
		actualDifficulty := normalizedDifficultyLabel(challenge.Difficulty)
		reasonCodes := append([]string(nil), target.ReasonCodes...)
		reasonCodes = append(reasonCodes, "difficulty_band_"+string(evaluation.RecommendedDifficultyBand))

		summary := recommendationReasonSummary(target, evaluation.RecommendedDifficultyBand, actualDifficulty)

		reasons = append(reasons, RecommendationReason{
			Dimension:      target.Dimension,
			DifficultyBand: evaluation.RecommendedDifficultyBand,
			Severity:       target.Severity,
			ReasonCodes:    reasonCodes,
			Summary:        summary,
			Evidence:       target.Evidence,
		})
	}

	return RecommendationPlan{
		WeakDimensions: evaluation.WeakDimensions,
		Targets:        evaluation.RecommendationTargets,
		Reasons:        reasons,
		DifficultyBand: evaluation.RecommendedDifficultyBand,
	}
}

func recommendationReasonSummary(
	target DimensionAdvice,
	preferred DifficultyBand,
	actualDifficulty string,
) string {
	label := dimensionLabel(target.Dimension)
	preferredDifficulty := string(preferred)
	actualMatchesPreferred := actualDifficulty == "" || strings.EqualFold(actualDifficulty, preferredDifficulty)

	if target.IsWeak {
		if actualMatchesPreferred {
			displayDifficulty := preferredDifficulty
			if actualDifficulty != "" {
				displayDifficulty = actualDifficulty
			}
			return fmt.Sprintf("%s 维度已经出现高置信度薄弱信号，这道 %s 难度题适合先补基础。", label, displayDifficulty)
		}
		return fmt.Sprintf("%s 维度已经出现高置信度薄弱信号，当前更适合先补 %s 难度；题库里这道 %s 难度题是最接近的候选。", label, preferredDifficulty, actualDifficulty)
	}

	if containsReasonCode(target.ReasonCodes, "progression_ready") {
		if actualMatchesPreferred {
			displayDifficulty := preferredDifficulty
			if actualDifficulty != "" {
				displayDifficulty = actualDifficulty
			}
			return fmt.Sprintf("%s 维度的基础已经比较稳定，这道 %s 难度题可以作为下一步进阶训练。", label, displayDifficulty)
		}
		return fmt.Sprintf("%s 维度的基础已经比较稳定，当前可以开始 %s 难度训练；题库里这道 %s 难度题是最接近的进阶候选。", label, preferredDifficulty, actualDifficulty)
	}

	if target.EvidenceCount < 2 {
		if actualMatchesPreferred {
			if actualDifficulty == "" {
				return fmt.Sprintf("%s 维度的训练证据还不够，这道题适合先补一条可靠样本。", label)
			}
			return fmt.Sprintf("%s 维度的训练证据还不够，这道 %s 难度题适合先补一条可靠样本。", label, actualDifficulty)
		}
		return fmt.Sprintf("%s 维度的训练证据还不够，当前更适合先补 %s 难度；题库里这道 %s 难度题可以先作为最接近的样本。", label, preferredDifficulty, actualDifficulty)
	}

	if actualMatchesPreferred {
		displayDifficulty := preferredDifficulty
		if actualDifficulty != "" {
			displayDifficulty = actualDifficulty
		}
		return fmt.Sprintf("%s 维度当前更适合先做 %s 难度训练，这道题可以作为下一步补强。", label, displayDifficulty)
	}
	return fmt.Sprintf("%s 维度当前更适合先做 %s 难度训练；题库里这道 %s 难度题可以作为最接近的补强候选。", label, preferredDifficulty, actualDifficulty)
}

func normalizedDifficultyLabel(difficulty string) string {
	return strings.ToLower(strings.TrimSpace(difficulty))
}

func recommendationTargetDimension(challenge ChallengeCandidate) string {
	dimension := strings.ToLower(strings.TrimSpace(challenge.Dimension))
	if dimension != "" {
		return dimension
	}
	return strings.ToLower(strings.TrimSpace(challenge.Category))
}

func evaluateDimension(fact DimensionFact) DimensionAdvice {
	dimension := strings.ToLower(strings.TrimSpace(fact.Dimension))
	if dimension == "" {
		return DimensionAdvice{}
	}

	profileScore := clampScore(fact.ProfileScore)
	attemptCount := maxInt(fact.AttemptCount, fact.SuccessCount)
	successCount := minInt(fact.SuccessCount, attemptCount)
	evidenceCount := fact.EvidenceCount
	if evidenceCount < attemptCount {
		evidenceCount = attemptCount
	}

	advice := DimensionAdvice{
		Dimension:     dimension,
		ProfileScore:  profileScore,
		AttemptCount:  attemptCount,
		SuccessCount:  successCount,
		EvidenceCount: evidenceCount,
	}

	successRate := 0.0
	if attemptCount > 0 {
		successRate = float64(successCount) / float64(attemptCount)
	}
	stableProgress := attemptCount >= 2 && successCount >= 2 && successRate >= 0.8

	reasonCodes := make([]string, 0, 3)
	if evidenceCount >= 4 && attemptCount >= 4 && profileScore < 0.35 && successCount == 0 {
		advice.IsWeak = true
		advice.Severity = SeverityDanger
		reasonCodes = append(reasonCodes, "score_critical", "evidence_sufficient")
	} else if evidenceCount >= 3 && attemptCount >= 3 && profileScore < 0.5 && successRate < 0.5 && !stableProgress {
		advice.IsWeak = true
		advice.Severity = SeverityWarning
		reasonCodes = append(reasonCodes, "score_low", "evidence_sufficient")
	} else if (evidenceCount > 0 || attemptCount > 0) && profileScore < 0.6 {
		advice.Severity = SeverityAttention
		if stableProgress {
			reasonCodes = append(reasonCodes, "coverage_gap", "recent_progress_stable")
		} else if successCount > 0 {
			reasonCodes = append(reasonCodes, "coverage_gap", "early_success_seen")
		} else if attemptCount >= 2 {
			reasonCodes = append(reasonCodes, "evidence_in_progress", "needs_foundation")
		} else {
			reasonCodes = append(reasonCodes, "evidence_insufficient", "needs_foundation")
		}
	} else {
		advice.Severity = SeverityGood
		reasonCodes = append(reasonCodes, "progress_stable")
	}

	if attemptCount >= 4 && successCount == 0 && advice.IsWeak {
		advice.Severity = maxSeverity(advice.Severity, SeverityDanger)
		reasonCodes = append(reasonCodes, "repeated_failures")
	}

	confidence := 0.2 + float64(minInt(evidenceCount, 4))*0.18
	if advice.IsWeak {
		confidence += 0.2
	}
	if successCount > 0 {
		confidence += 0.05
	}
	advice.Confidence = roundConfidence(confidence)
	advice.ReasonCodes = uniqueStrings(reasonCodes)
	advice.Summary = buildDimensionSummary(advice)
	advice.Evidence = fmt.Sprintf("%s 维度画像 %.0f%%，尝试 %d 次，成功 %d 次，相关证据 %d 条。", dimensionLabel(dimension), profileScore*100, attemptCount, successCount, evidenceCount)
	return advice
}

func buildDimensionSummary(advice DimensionAdvice) string {
	switch {
	case advice.IsWeak && advice.Severity == SeverityDanger:
		return fmt.Sprintf("%s 维度得分偏低，而且已经有足够训练证据支撑弱项判断。", dimensionLabel(advice.Dimension))
	case advice.IsWeak:
		return fmt.Sprintf("%s 维度经过多次训练后仍然偏弱，已经接近真实薄弱项。", dimensionLabel(advice.Dimension))
	case containsReasonCode(advice.ReasonCodes, "progression_ready"):
		return fmt.Sprintf("%s 维度当前基础比较稳定，可以开始做更高一档的进阶训练。", dimensionLabel(advice.Dimension))
	case containsReasonCode(advice.ReasonCodes, "recent_progress_stable"):
		return fmt.Sprintf("%s 维度已经开始出现稳定进展，但覆盖率还不够，建议继续补样本。", dimensionLabel(advice.Dimension))
	case containsReasonCode(advice.ReasonCodes, "early_success_seen"):
		return fmt.Sprintf("%s 维度已经出现成功样本，但覆盖率还不够，暂时更像补样本而不是明确弱项。", dimensionLabel(advice.Dimension))
	case containsReasonCode(advice.ReasonCodes, "evidence_in_progress"):
		return fmt.Sprintf("%s 维度已经出现几次训练样本，但结果还不稳定，建议先补基础再判断。", dimensionLabel(advice.Dimension))
	case advice.Severity == SeverityAttention:
		return fmt.Sprintf("%s 维度暂时还缺少足够训练样本，建议先补基础证据。", dimensionLabel(advice.Dimension))
	default:
		return fmt.Sprintf("%s 维度当前没有明显风险。", dimensionLabel(advice.Dimension))
	}
}

func pickRecommendationTarget(category string, evaluation StudentEvaluation) DimensionAdvice {
	normalized := strings.ToLower(strings.TrimSpace(category))
	for _, item := range evaluation.RecommendationTargets {
		if item.Dimension == normalized {
			return item
		}
	}
	if len(evaluation.RecommendationTargets) > 0 {
		return evaluation.RecommendationTargets[0]
	}
	return DimensionAdvice{}
}

func selectProgressionTarget(facts []DimensionFact, dimensions []DimensionAdvice) (DimensionAdvice, DifficultyBand) {
	type candidate struct {
		advice DimensionAdvice
		band   DifficultyBand
	}

	adviceByDimension := make(map[string]DimensionAdvice, len(dimensions))
	for _, item := range dimensions {
		adviceByDimension[item.Dimension] = item
	}

	candidates := make([]candidate, 0, len(facts))
	for _, fact := range facts {
		dimension := strings.ToLower(strings.TrimSpace(fact.Dimension))
		if dimension == "" {
			continue
		}
		advice, ok := adviceByDimension[dimension]
		if !ok {
			continue
		}
		band, ready := progressionBandForFact(advice, fact)
		if !ready {
			continue
		}
		advice.ReasonCodes = uniqueStrings(append(append([]string(nil), advice.ReasonCodes...), "progression_ready"))
		advice.Summary = buildDimensionSummary(advice)
		candidates = append(candidates, candidate{advice: advice, band: band})
	}

	if len(candidates) == 0 {
		return DimensionAdvice{}, ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		if difficultyBandRank(candidates[i].band) != difficultyBandRank(candidates[j].band) {
			return difficultyBandRank(candidates[i].band) > difficultyBandRank(candidates[j].band)
		}
		if candidates[i].advice.ProfileScore != candidates[j].advice.ProfileScore {
			return candidates[i].advice.ProfileScore > candidates[j].advice.ProfileScore
		}
		if candidates[i].advice.SuccessCount != candidates[j].advice.SuccessCount {
			return candidates[i].advice.SuccessCount > candidates[j].advice.SuccessCount
		}
		return candidates[i].advice.Dimension < candidates[j].advice.Dimension
	})

	return candidates[0].advice, candidates[0].band
}

func progressionBandForFact(advice DimensionAdvice, fact DimensionFact) (DifficultyBand, bool) {
	if advice.Severity != SeverityGood {
		return "", false
	}
	if advice.ProfileScore < 0.75 || advice.SuccessCount < 2 || advice.EvidenceCount < 4 {
		return "", false
	}
	solved := fact.SolvedDifficultyCounts
	if len(solved) == 0 {
		return "", false
	}

	switch {
	case solved[model.ChallengeDifficultyHard] > 0 && solved[model.ChallengeDifficultyMedium] > 0:
		return DifficultyBandInsane, true
	case solved[model.ChallengeDifficultyMedium] > 0 && solved[model.ChallengeDifficultyEasy] > 0:
		return DifficultyBandHard, true
	case solved[model.ChallengeDifficultyEasy] > 0:
		return DifficultyBandMedium, true
	case solved[model.ChallengeDifficultyBeginner] > 0:
		return DifficultyBandEasy, true
	default:
		return "", false
	}
}

func difficultyBandRank(band DifficultyBand) int {
	switch band {
	case DifficultyBandBeginner:
		return 1
	case DifficultyBandEasy:
		return 2
	case DifficultyBandMedium:
		return 3
	case DifficultyBandHard:
		return 4
	case DifficultyBandInsane:
		return 5
	default:
		return 0
	}
}

func selectTopWeakDimension(
	counts map[string]int,
	grouped map[string][]StudentFactSnapshot,
	totalStudents int,
) (string, []StudentFactSnapshot) {
	bestDimension := ""
	bestCount := 0
	for dimension, count := range counts {
		if count > bestCount || (count == bestCount && (bestDimension == "" || dimension < bestDimension)) {
			bestDimension = dimension
			bestCount = count
		}
	}
	if bestDimension == "" {
		return "", nil
	}
	minClusterCount := maxInt(2, (totalStudents+3)/4)
	if bestCount < minClusterCount {
		return "", nil
	}

	students := append([]StudentFactSnapshot(nil), grouped[bestDimension]...)
	sortStudentsByRisk(students)
	return bestDimension, students
}

func sortStudentsByRisk(students []StudentFactSnapshot) {
	sort.Slice(students, func(i, j int) bool {
		if students[i].MaxWrongStreak != students[j].MaxWrongStreak {
			return students[i].MaxWrongStreak > students[j].MaxWrongStreak
		}
		if students[i].RecentEventCount7d != students[j].RecentEventCount7d {
			return students[i].RecentEventCount7d < students[j].RecentEventCount7d
		}
		if students[i].CorrectSubmissionCount != students[j].CorrectSubmissionCount {
			return students[i].CorrectSubmissionCount < students[j].CorrectSubmissionCount
		}
		return students[i].Username < students[j].Username
	})
}

func classDifficultyBand(
	students []StudentFactSnapshot,
	evaluations map[int64]StudentEvaluation,
	dimension string,
) DifficultyBand {
	best := DifficultyBandEasy
	for _, student := range students {
		evaluation, ok := evaluations[student.UserID]
		if !ok {
			continue
		}
		for _, item := range evaluation.RecommendationTargets {
			if item.Dimension != dimension {
				continue
			}
			if evaluation.RecommendedDifficultyBand == DifficultyBandBeginner {
				return DifficultyBandBeginner
			}
			best = evaluation.RecommendedDifficultyBand
		}
	}
	return best
}

func submissionStabilitySeverity(snapshot StudentFactSnapshot) (Severity, bool) {
	wrongCount := submissionFailureCount(snapshot)
	successCount := submissionSuccessCount(snapshot)
	if snapshot.MaxWrongStreak >= 4 || (wrongCount >= 5 && wrongCount > successCount*2) {
		return SeverityDanger, true
	}
	if snapshot.MaxWrongStreak >= 2 || (wrongCount >= 3 && wrongCount > successCount) {
		return SeverityWarning, true
	}
	return "", false
}

func buildLowActivityObservation(snapshot StudentFactSnapshot) (ReviewArchiveObservation, bool) {
	severity, ok := lowActivitySeverity(snapshot)
	if !ok {
		return ReviewArchiveObservation{}, false
	}

	observation := ReviewArchiveObservation{
		Code:     "low_activity",
		Label:    "近期活跃度",
		Severity: severity,
		Summary:  "最近一周训练节奏偏慢，需要尽快恢复稳定投入。",
		Evidence: fmt.Sprintf("近 7 天活跃 %d 天，训练事件 %d 次，成功事件 %d 次。", snapshot.ActiveDays7d, snapshot.RecentEventCount7d, submissionSuccessCount(snapshot)),
		Action:   "先确认当前卡点，再安排 1 个可完成的小目标，把训练节奏拉回来。",
	}
	if severity == SeverityWarning {
		observation.Summary = "最近一周训练投入明显不足，已经有掉队风险。"
		observation.Action = "优先确认是否卡题或脱节，并安排一次短时补练恢复训练节奏。"
	}
	return observation, true
}

func lowActivitySeverity(snapshot StudentFactSnapshot) (Severity, bool) {
	if !isLowActivity(snapshot) {
		return "", false
	}
	if snapshot.ActiveDays7d <= 1 && snapshot.RecentEventCount7d <= 2 {
		return SeverityWarning, true
	}
	return SeverityAttention, true
}

func isLowActivity(snapshot StudentFactSnapshot) bool {
	if snapshot.ActiveDays7d <= 1 {
		return true
	}
	if snapshot.RecentEventCount7d <= 2 && submissionSuccessCount(snapshot) == 0 {
		return true
	}
	return false
}

func buildTrainingClosureEvidence(
	challengeSuccessCount int,
	submissionSuccessCount int,
	awdSuccessCount int,
	writeupCount int,
	approvedReviewCount int,
) string {
	parts := []string{
		fmt.Sprintf("完成题目 %d 道", challengeSuccessCount),
		fmt.Sprintf("writeup %d 份", writeupCount),
		fmt.Sprintf("通过评阅 %d 条", approvedReviewCount),
	}
	if awdSuccessCount > 0 {
		parts = append(parts, fmt.Sprintf("AWD 成功 %d 次", awdSuccessCount))
	}
	if submissionSuccessCount > challengeSuccessCount && awdSuccessCount == 0 {
		parts = append(parts, fmt.Sprintf("归档成功事件 %d 次", submissionSuccessCount))
	}
	return strings.Join(parts, "，") + "。"
}

func buildWeakDimensionObservationSummary(advice DimensionAdvice) string {
	if advice.Severity == SeverityDanger {
		return fmt.Sprintf("%s 维度已经形成高置信度薄弱信号。", dimensionLabel(advice.Dimension))
	}
	if advice.SuccessCount > 0 {
		return fmt.Sprintf("%s 维度已经有过成功样本，但多次训练后结果仍不稳定，开始接近真实薄弱项。", dimensionLabel(advice.Dimension))
	}
	return fmt.Sprintf("%s 维度经过多次训练后仍然偏弱，已经形成明确补强优先级。", dimensionLabel(advice.Dimension))
}

func buildCoverageGapObservationSummary(advice DimensionAdvice) string {
	switch {
	case advice.SuccessCount > 0:
		return fmt.Sprintf("%s 维度已经出现成功样本，但覆盖率还不够，暂时不宜下明确弱项结论。", dimensionLabel(advice.Dimension))
	case advice.AttemptCount >= 2:
		return fmt.Sprintf("%s 维度已经出现少量训练样本，但结果还不稳定，暂时不宜下明确弱项结论。", dimensionLabel(advice.Dimension))
	default:
		return fmt.Sprintf("%s 维度的训练证据还不够，暂时不宜下明确弱项结论。", dimensionLabel(advice.Dimension))
	}
}

func challengeSuccessCount(snapshot StudentFactSnapshot) int {
	if snapshot.ChallengeSuccessCount > 0 {
		return snapshot.ChallengeSuccessCount
	}
	if hasExplicitSubmissionCounts(snapshot) {
		return maxInt(snapshot.SubmissionSuccessCount-snapshot.AWDSuccessCount, 0)
	}
	return snapshot.CorrectSubmissionCount
}

func submissionSuccessCount(snapshot StudentFactSnapshot) int {
	if hasExplicitSubmissionCounts(snapshot) {
		return snapshot.SubmissionSuccessCount
	}
	return snapshot.CorrectSubmissionCount
}

func submissionFailureCount(snapshot StudentFactSnapshot) int {
	if hasExplicitSubmissionCounts(snapshot) {
		return snapshot.SubmissionFailureCount
	}
	return snapshot.WrongSubmissionCount
}

func hasExplicitSubmissionCounts(snapshot StudentFactSnapshot) bool {
	return snapshot.ChallengeSuccessCount > 0 || snapshot.SubmissionSuccessCount > 0 || snapshot.SubmissionFailureCount > 0
}

func studentIDs(students []StudentFactSnapshot, limit int) []int64 {
	if limit <= 0 || len(students) < limit {
		limit = len(students)
	}
	ids := make([]int64, 0, limit)
	for _, student := range students[:limit] {
		ids = append(ids, student.UserID)
	}
	return ids
}

func firstStudentID(students []StudentFactSnapshot) *int64 {
	if len(students) == 0 {
		return nil
	}
	id := students[0].UserID
	return &id
}

func dimensionLabel(dimension string) string {
	switch strings.ToLower(strings.TrimSpace(dimension)) {
	case "web":
		return "Web"
	case "pwn":
		return "Pwn"
	case "reverse":
		return "逆向"
	case "crypto":
		return "密码"
	case "misc":
		return "杂项"
	case "forensics":
		return "取证"
	default:
		return strings.ToUpper(strings.TrimSpace(dimension))
	}
}

func maxSeverity(left, right Severity) Severity {
	if severityRank(right) > severityRank(left) {
		return right
	}
	return left
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityDanger:
		return 4
	case SeverityWarning:
		return 3
	case SeverityAttention:
		return 2
	case SeverityGood:
		return 1
	default:
		return 0
	}
}

func roundConfidence(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return float64(int(value*100+0.5)) / 100
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

func containsReasonCode(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}

func clampScore(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
