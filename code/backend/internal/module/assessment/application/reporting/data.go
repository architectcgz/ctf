package reporting

import (
	"time"

	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

type PersonalReportData struct {
	User           *assessmentdomain.ReportUser
	SkillProfile   []*assessmentcontracts.SkillDimension
	Stats          *assessmentdomain.PersonalReportStats
	DimensionStats []assessmentdomain.ReportDimensionStat
}

type ClassReportData struct {
	ClassName              string                                        `json:"class_name"`
	Window                 ClassReportWindow                             `json:"window"`
	TotalStudents          int                                           `json:"total_students"`
	AverageScore           float64                                       `json:"average_score"`
	DimensionAverages      []assessmentdomain.ClassDimensionAverage      `json:"dimension_averages"`
	TopStudents            []assessmentdomain.ClassTopStudent            `json:"top_students"`
	Summary                *ClassReportSummary                           `json:"summary,omitempty"`
	Trend                  *ClassReportTrend                             `json:"trend,omitempty"`
	Review                 *ClassReportReview                            `json:"review,omitempty"`
	CategoryDistribution   []assessmentdomain.ClassDistributionStat      `json:"category_distribution"`
	DifficultyDistribution []assessmentdomain.ClassDistributionStat      `json:"difficulty_distribution"`
	ContestMigration       assessmentdomain.ClassContestMigrationSummary `json:"contest_migration"`
}

type ClassReportWindow struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	Days     int    `json:"days"`
}

type ClassReportSummary struct {
	ClassName          string  `json:"class_name"`
	StudentCount       int64   `json:"student_count"`
	AverageSolved      float64 `json:"average_solved"`
	ActiveStudentCount int64   `json:"active_student_count"`
	ActiveRate         float64 `json:"active_rate"`
	RecentEventCount   int64   `json:"recent_event_count"`
}

type ClassReportTrendPoint struct {
	Date               string `json:"date"`
	ActiveStudentCount int64  `json:"active_student_count"`
	EventCount         int64  `json:"event_count"`
	SolveCount         int64  `json:"solve_count"`
}

type ClassReportTrend struct {
	ClassName string                  `json:"class_name"`
	Points    []ClassReportTrendPoint `json:"points"`
}

type ClassReportReviewStudentRef struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Name     *string `json:"name,omitempty"`
}

type ClassReportRecommendationItem struct {
	ChallengeID    int64    `json:"challenge_id"`
	Title          string   `json:"title"`
	Category       string   `json:"category"`
	Difficulty     string   `json:"difficulty"`
	Dimension      string   `json:"dimension,omitempty"`
	DifficultyBand string   `json:"difficulty_band,omitempty"`
	Severity       string   `json:"severity,omitempty"`
	ReasonCodes    []string `json:"reason_codes,omitempty"`
	Summary        string   `json:"summary"`
	Evidence       string   `json:"evidence,omitempty"`
}

type ClassReportReviewItem struct {
	Code           string                         `json:"code"`
	Severity       string                         `json:"severity"`
	Summary        string                         `json:"summary"`
	Evidence       string                         `json:"evidence,omitempty"`
	Action         string                         `json:"action,omitempty"`
	ReasonCodes    []string                       `json:"reason_codes,omitempty"`
	Dimension      string                         `json:"dimension,omitempty"`
	Students       []ClassReportReviewStudentRef  `json:"students,omitempty"`
	Recommendation *ClassReportRecommendationItem `json:"recommendation,omitempty"`
}

type ClassReportReview struct {
	ClassName string                  `json:"class_name"`
	Items     []ClassReportReviewItem `json:"items"`
}

type ContestExportData struct {
	GeneratedAt time.Time                                      `json:"generated_at"`
	Contest     ContestExportMeta                              `json:"contest"`
	Scoreboard  []assessmentdomain.ContestExportScoreboardItem `json:"scoreboard"`
	Challenges  []assessmentdomain.ContestExportChallengeItem  `json:"challenges"`
	Teams       []assessmentdomain.ContestExportTeamItem       `json:"teams"`
}

type ContestExportMeta struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Mode        string     `json:"mode"`
	Status      string     `json:"status"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	FreezeTime  *time.Time `json:"freeze_time,omitempty"`
}

type ReviewArchiveData struct {
	GeneratedAt         time.Time                                         `json:"generated_at"`
	Student             ReviewArchiveStudent                              `json:"student"`
	Summary             assessmentdomain.ReviewArchiveSummary             `json:"summary"`
	SkillProfile        []*assessmentcontracts.SkillDimension             `json:"skill_profile,omitempty"`
	Timeline            []assessmentdomain.ReviewArchiveTimelineEvent     `json:"timeline"`
	Evidence            []assessmentdomain.ReviewArchiveEvidenceEvent     `json:"evidence"`
	Writeups            []assessmentdomain.ReviewArchiveWriteupItem       `json:"writeups"`
	ManualReviews       []assessmentdomain.ReviewArchiveManualReviewItem  `json:"manual_reviews"`
	TeacherObservations assessmentdomain.ReviewArchiveTeacherObservations `json:"teacher_observations"`
}

type ReviewArchiveStudent struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	ClassName string `json:"class_name,omitempty"`
}
