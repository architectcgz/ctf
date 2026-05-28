package main

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"ctf-platform/internal/config"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingqueries "ctf-platform/internal/module/teaching_query/application/queries"
	"ctf-platform/internal/shared/taxonomy"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildCoverageStudentScenariosSkipsSmallCatalog(t *testing.T) {
	catalog := newSyntheticChallengeCatalog(t, 4)

	scenarios := buildCoverageStudentScenarios(catalog)
	if len(scenarios) != 0 {
		t.Fatalf("expected no coverage scenarios for small catalog, got %d", len(scenarios))
	}
}

func TestBuildCoverageStudentScenariosExpandsLargeCatalogAcrossDimensions(t *testing.T) {
	catalog := newSyntheticChallengeCatalog(t, 14)

	scenarios := buildCoverageStudentScenarios(catalog)
	if len(scenarios) == 0 {
		t.Fatal("expected coverage scenarios for large catalog")
	}

	weakCount := make(map[string]int, len(taxonomy.AllDimensions))
	weakChallengeUsage := make(map[string]map[int]struct{}, len(taxonomy.AllDimensions))
	for _, scenario := range scenarios {
		weakDimension := weakestScenarioDimension(scenario)
		if weakDimension == "" {
			t.Fatalf("scenario %q does not expose a weak dimension", scenario.Label)
		}
		weakCount[weakDimension]++
		if weakChallengeUsage[weakDimension] == nil {
			weakChallengeUsage[weakDimension] = make(map[int]struct{})
		}
		for _, session := range scenario.Sessions {
			if session.ChallengeCategory != weakDimension {
				continue
			}
			weakChallengeUsage[weakDimension][session.ChallengeIndex] = struct{}{}
		}
	}

	for _, dimension := range taxonomy.AllDimensions {
		if weakCount[dimension] == 0 {
			t.Fatalf("expected coverage scenarios for dimension %s, got none", dimension)
		}
		if len(weakChallengeUsage[dimension]) < 2 {
			t.Fatalf("expected weak challenge usage for dimension %s to span multiple challenge indexes, got %d", dimension, len(weakChallengeUsage[dimension]))
		}
	}
}

func TestBuildStudentScenariosIncludesBaseAndCoverageForLargeCatalog(t *testing.T) {
	catalog := newSyntheticChallengeCatalog(t, 14)

	scenarios := buildStudentScenarios(catalog)
	if len(scenarios) <= 7 {
		t.Fatalf("expected large catalog to expand beyond 7 base scenarios, got %d", len(scenarios))
	}

	foundBase := false
	foundCoverage := false
	for _, scenario := range scenarios {
		switch scenario.Label {
		case "稳定闭环 + AWD 迁移":
			foundBase = true
		}
		if len(scenario.Label) >= len("扩展覆盖") && scenario.Label[:len("扩展覆盖")] == "扩展覆盖" {
			foundCoverage = true
		}
	}
	if !foundBase {
		t.Fatal("expected baseline scenario to remain present")
	}
	if !foundCoverage {
		t.Fatal("expected expanded catalog scenario to be present")
	}
}

func TestBuildCoverageStudentScenariosDistributesTopRecommendationCandidates(t *testing.T) {
	catalog := newSyntheticChallengeCatalog(t, 14)

	scenarios := buildCoverageStudentScenarios(catalog)
	topRecommendationIDs := make(map[string]map[int64]struct{}, len(taxonomy.AllDimensions))
	for _, scenario := range scenarios {
		weakDimension := weakestScenarioDimension(scenario)
		if weakDimension == "" {
			t.Fatalf("scenario %q does not expose a weak dimension", scenario.Label)
		}
		topChallengeID := syntheticTopRecommendationChallengeID(t, catalog, scenario, weakDimension)
		if topRecommendationIDs[weakDimension] == nil {
			topRecommendationIDs[weakDimension] = make(map[int64]struct{})
		}
		topRecommendationIDs[weakDimension][topChallengeID] = struct{}{}
	}

	expectedPerDimension := coverageStudentsPerCategory(14)
	for _, dimension := range taxonomy.AllDimensions {
		if got := len(topRecommendationIDs[dimension]); got != expectedPerDimension {
			t.Fatalf("expected %d distinct top recommendation candidates for %s, got %d", expectedPerDimension, dimension, got)
		}
	}
}

func TestBuildCoverageStudentScenariosMixesRetryPressure(t *testing.T) {
	catalog := newSyntheticChallengeCatalog(t, 14)

	scenarios := buildCoverageStudentScenarios(catalog)
	highRiskByDimension := make(map[string]bool, len(taxonomy.AllDimensions))
	stableByDimension := make(map[string]bool, len(taxonomy.AllDimensions))
	for _, scenario := range scenarios {
		weakDimension := weakestScenarioDimension(scenario)
		if weakDimension == "" {
			t.Fatalf("scenario %q does not expose a weak dimension", scenario.Label)
		}
		if syntheticMaxWrongStreak(scenario) >= 3 {
			highRiskByDimension[weakDimension] = true
			continue
		}
		stableByDimension[weakDimension] = true
	}

	for _, dimension := range taxonomy.AllDimensions {
		if !highRiskByDimension[dimension] {
			t.Fatalf("expected at least one high-risk retry scenario for %s", dimension)
		}
		if !stableByDimension[dimension] {
			t.Fatalf("expected at least one non-streak-heavy scenario for %s", dimension)
		}
	}
}

func TestBuildCoverageStudentScenariosFallsBackToAvailableSupportDimensions(t *testing.T) {
	t.Parallel()

	catalog := newSparseSyntheticChallengeCatalog(t, map[string]int{
		taxonomy.DimensionWeb:     24,
		taxonomy.DimensionReverse: 24,
	})

	scenarios := buildCoverageStudentScenarios(catalog)
	if len(scenarios) == 0 {
		t.Fatal("expected coverage scenarios for sparse large catalog")
	}

	for _, scenario := range scenarios {
		for _, session := range scenario.Sessions {
			if len(catalog.byCategory[session.ChallengeCategory]) == 0 {
				t.Fatalf("scenario %q used unavailable category %s", scenario.Label, session.ChallengeCategory)
			}
		}
	}
}

func TestBuildSeedCoverageSummaryReportsPublishedUsedAndRecommendationReach(t *testing.T) {
	catalog := newSyntheticChallengeCatalog(t, 14)
	scenarios := buildStudentScenarios(catalog)
	results := []seededStudentResult{
		{
			Recommendations: teachingqueries.TeacherRecommendationResp{
				Challenges: []teachingqueries.TeacherRecommendationItem{
					{ChallengeID: 101, Title: "web-1"},
				},
			},
		},
		{
			Recommendations: teachingqueries.TeacherRecommendationResp{
				Challenges: []teachingqueries.TeacherRecommendationItem{
					{ChallengeID: 101, Title: "web-1"},
				},
			},
		},
		{
			Recommendations: teachingqueries.TeacherRecommendationResp{
				Challenges: []teachingqueries.TeacherRecommendationItem{
					{ChallengeID: 202, Title: "pwn-2"},
				},
			},
		},
	}

	summary := buildSeedCoverageSummary(catalog, scenarios, results)
	expectedPublished := len(taxonomy.AllDimensions) * 14
	if summary.PublishedChallenges != expectedPublished {
		t.Fatalf("expected published challenge count %d, got %d", expectedPublished, summary.PublishedChallenges)
	}
	if summary.UsedPracticeChallenges <= 7 {
		t.Fatalf("expected summary to report wider practice coverage, got %d", summary.UsedPracticeChallenges)
	}
	if summary.StudentsWithRecommendations != 3 {
		t.Fatalf("expected 3 students with recommendations, got %d", summary.StudentsWithRecommendations)
	}
	if summary.UniqueTopRecommendationCount != 2 {
		t.Fatalf("expected 2 unique top recommendations, got %d", summary.UniqueTopRecommendationCount)
	}
	for _, dimension := range taxonomy.AllDimensions {
		category, ok := summary.ByCategory[dimension]
		if !ok {
			t.Fatalf("expected category summary for %s", dimension)
		}
		if category.Published != 14 {
			t.Fatalf("expected published count 14 for %s, got %d", dimension, category.Published)
		}
		if category.Used == 0 {
			t.Fatalf("expected used count > 0 for %s", dimension)
		}
	}
}

func TestBuildBaseStudentScenariosIncludeRichAWDReviewData(t *testing.T) {
	t.Parallel()

	var awd *awdScenario
	for _, scenario := range buildBaseStudentScenarios() {
		if scenario.AWD == nil {
			continue
		}
		awd = scenario.AWD
		break
	}
	if awd == nil {
		t.Fatal("expected at least one base awd scenario")
	}
	if len(awd.Teams) != 3 {
		t.Fatalf("expected 3 awd teams, got %d", len(awd.Teams))
	}
	if len(awd.Rounds) != 3 {
		t.Fatalf("expected 3 awd rounds, got %d", len(awd.Rounds))
	}

	totalServices := 0
	totalTraffic := 0
	for _, round := range awd.Rounds {
		if len(round.Services) != 3 {
			t.Fatalf("expected 3 services per round, got %d in round %d", len(round.Services), round.RoundNumber)
		}
		if len(round.Traffic) != 3 {
			t.Fatalf("expected 3 traffic events per round, got %d in round %d", len(round.Traffic), round.RoundNumber)
		}
		if len(round.Attacks) == 0 {
			t.Fatalf("expected attacks in round %d", round.RoundNumber)
		}
		totalServices += len(round.Services)
		totalTraffic += len(round.Traffic)
	}
	if totalServices != 9 {
		t.Fatalf("expected 9 awd services in total, got %d", totalServices)
	}
	if totalTraffic != 9 {
		t.Fatalf("expected 9 awd traffic events in total, got %d", totalTraffic)
	}
	if countAWDAttacks(awd) != 7 {
		t.Fatalf("expected 7 awd attacks in total, got %d", countAWDAttacks(awd))
	}
}

func TestBuildBaseStudentScenariosCoverAWDOutcomeEdges(t *testing.T) {
	t.Parallel()

	hasSuccess := false
	hasRepeatedFailure := false
	hasSingleProbe := false

	for _, scenario := range buildBaseStudentScenarios() {
		if scenario.AWD == nil {
			continue
		}
		submittedByStudent := countStudentSubmittedAWDAttacks(scenario.AWD)
		successCount := 0
		for _, round := range scenario.AWD.Rounds {
			for _, attack := range round.Attacks {
				if attack.SubmittedByStudent && attack.IsSuccess && attack.ScoreGained > 0 {
					successCount++
				}
			}
		}
		switch {
		case successCount > 0:
			hasSuccess = true
		case submittedByStudent >= 2:
			hasRepeatedFailure = true
		case submittedByStudent == 1:
			hasSingleProbe = true
		}
	}

	if !hasSuccess || !hasRepeatedFailure || !hasSingleProbe {
		t.Fatalf("expected AWD seed matrix to cover success/repeated-failure/single-probe edges, got success=%t repeatedFailure=%t singleProbe=%t", hasSuccess, hasRepeatedFailure, hasSingleProbe)
	}
}

func TestSeedStudentAWDScenarioBuildsTeacherReviewArchiveEvidence(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupAWDTestDB(t)
	now := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	teacher := createSeedReviewUser(t, db, 5101, seedTeacherUsername, "赵晓峰", identitycontracts.RoleTeacher, now)
	student := createSeedReviewUser(t, db, 5102, "linchenxi", "林宸熙", identitycontracts.RoleStudent, now)

	if err := db.Create(&challengeentity.AWDChallenge{
		ID:             91001,
		Name:           "awd-web-seed",
		Slug:           "awd-web-seed",
		Category:       "web",
		Difficulty:     challengeentity.ChallengeDifficultyMedium,
		ServiceType:    challengeentity.AWDServiceTypeWebHTTP,
		DeploymentMode: challengeentity.AWDDeploymentModeSingleContainer,
		Status:         challengeentity.AWDChallengeStatusPublished,
		CheckerType:    contestcontracts.AWDCheckerTypeHTTPStandard,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	awdCatalog, err := loadAWDChallengeCatalog(context.Background(), db)
	if err != nil {
		t.Fatalf("load awd catalog: %v", err)
	}

	var awdSeed *awdScenario
	for _, scenario := range buildBaseStudentScenarios() {
		if scenario.User.Username != student.Username || scenario.AWD == nil {
			continue
		}
		awdSeed = scenario.AWD
		break
	}
	if awdSeed == nil {
		t.Fatal("expected base awd seed for linchenxi")
	}

	if err := seedStudentAWDScenario(db, teacher, student, awdSeed, awdCatalog, now); err != nil {
		t.Fatalf("seed awd scenario: %v", err)
	}

	var contest contestcontracts.Contest
	if err := db.Where("mode = ? AND title LIKE ?", contestcontracts.ContestModeAWD, seedAWDContestTitle+"%").First(&contest).Error; err != nil {
		t.Fatalf("load seeded awd contest: %v", err)
	}

	var memberCount int64
	if err := db.Model(&contestcontracts.TeamMember{}).Where("contest_id = ?", contest.ID).Count(&memberCount).Error; err != nil {
		t.Fatalf("count team members: %v", err)
	}
	if memberCount != 7 {
		t.Fatalf("expected 7 team members, got %d", memberCount)
	}

	var serviceCount int64
	if err := db.Model(&contestcontracts.AWDTeamService{}).Joins("JOIN awd_rounds ON awd_rounds.id = awd_team_services.round_id").Where("awd_rounds.contest_id = ?", contest.ID).Count(&serviceCount).Error; err != nil {
		t.Fatalf("count awd team services: %v", err)
	}
	if serviceCount != 9 {
		t.Fatalf("expected 9 awd team services, got %d", serviceCount)
	}

	var trafficCount int64
	if err := db.Model(&contestcontracts.AWDTrafficEvent{}).Where("contest_id = ?", contest.ID).Count(&trafficCount).Error; err != nil {
		t.Fatalf("count awd traffic events: %v", err)
	}
	if trafficCount != 9 {
		t.Fatalf("expected 9 awd traffic events, got %d", trafficCount)
	}

	service := assessmentqry.NewTeacherAWDReviewService(
		assessmentinfra.NewTeacherAWDReviewRepository(db),
		config.PaginationConfig{DefaultPageSize: 20, MaxPageSize: 100},
	)
	roundNumber := 2
	resp, err := service.GetContestArchive(context.Background(), teacher.ID, contest.ID, assessmentqry.GetTeacherAWDReviewArchiveInput{
		RoundNumber: &roundNumber,
	})
	if err != nil {
		t.Fatalf("GetContestArchive() error = %v", err)
	}
	if resp.Scope.SnapshotType != "final" {
		t.Fatalf("expected final snapshot, got %+v", resp.Scope)
	}
	if resp.Overview == nil {
		t.Fatalf("expected archive overview, got %+v", resp)
	}
	if resp.Overview.RoundCount != 3 || resp.Overview.TeamCount != 3 {
		t.Fatalf("unexpected overview counts: %+v", resp.Overview)
	}
	if resp.Overview.ServiceCount != 9 || resp.Overview.AttackCount != 7 || resp.Overview.TrafficCount != 9 {
		t.Fatalf("unexpected evidence counts: %+v", resp.Overview)
	}
	if resp.Contest.LatestEvidenceAt == nil {
		t.Fatalf("expected latest evidence timestamp, got %+v", resp.Contest)
	}
	if len(resp.Rounds) != 3 {
		t.Fatalf("expected 3 rounds, got %+v", resp.Rounds)
	}
	if resp.SelectedRound == nil || resp.SelectedRound.Round.RoundNumber != 2 {
		t.Fatalf("expected selected round 2, got %+v", resp.SelectedRound)
	}
	if len(resp.SelectedRound.Teams) != 3 {
		t.Fatalf("expected 3 teams in selected round, got %+v", resp.SelectedRound.Teams)
	}
	for _, team := range resp.SelectedRound.Teams {
		if team.MemberCount == 0 {
			t.Fatalf("expected member count for team %+v", team)
		}
	}
	if len(resp.SelectedRound.Services) != 3 || resp.SelectedRound.Round.ServiceCount != 3 {
		t.Fatalf("expected 3 selected-round services, got %+v", resp.SelectedRound.Services)
	}
	if len(resp.SelectedRound.Attacks) != 3 || resp.SelectedRound.Round.AttackCount != 3 {
		t.Fatalf("expected 3 selected-round attacks, got %+v", resp.SelectedRound.Attacks)
	}
	if len(resp.SelectedRound.Traffic) != 3 || resp.SelectedRound.Round.TrafficCount != 3 {
		t.Fatalf("expected 3 selected-round traffic events, got %+v", resp.SelectedRound.Traffic)
	}
}

func TestSeedStudentAWDScenarioCreatesContestAWDServiceParent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&contestcontracts.Contest{},
		&contestcontracts.ContestAWDService{},
		&contestcontracts.Team{},
		&contestcontracts.TeamMember{},
		&contestcontracts.AWDRound{},
		&contestcontracts.AWDTeamService{},
		&contestcontracts.AWDAttackLog{},
		&contestcontracts.AWDTrafficEvent{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	teacher := &identitycontracts.User{ID: 1001, Username: "teacher"}
	student := &identitycontracts.User{ID: 1002, Username: "student"}
	catalog := &awdChallengeCatalog{
		byCategory: map[string][]awdChallengeRef{
			"web": {
				{
					ID:               16,
					Name:             "样本 Web AWD",
					Category:         "web",
					Difficulty:       "medium",
					Description:      "seed test",
					ServiceType:      "web_http",
					DeploymentMode:   "single_container",
					CheckerType:      contestcontracts.AWDCheckerTypeHTTPStandard,
					CheckerConfig:    `{"method":"GET","path":"/health"}`,
					FlagMode:         "dynamic_per_round",
					FlagConfig:       `{"flag_prefix":"awd"}`,
					DefenseEntryMode: "none",
					AccessConfig:     `{"port":8080}`,
					RuntimeConfig:    `{"checker_token_env":"TOKEN"}`,
				},
			},
		},
	}
	scenario := &awdScenario{
		ChallengeCategory: "web",
		ChallengeIndex:    0,
		StartOffset:       -time.Hour,
		Teams: []awdTeamSeed{
			{Name: "红队", CaptainRole: seedAWDCaptainStudent, MemberCount: 2},
			{Name: "蓝队", CaptainRole: seedAWDCaptainTeacher, MemberCount: 2},
		},
		Rounds: []awdRoundSeed{
			{
				RoundNumber:  1,
				StartOffset:  0,
				Duration:     20 * time.Minute,
				AttackScore:  30,
				DefenseScore: 3,
				Services: []awdServiceSeed{
					{TeamIndex: 0, ServiceStatus: contestcontracts.AWDServiceStatusUp, SLAScore: 1, DefenseScore: 2, AttackScore: 0, CheckerType: contestcontracts.AWDCheckerTypeHTTPStandard},
					{TeamIndex: 1, ServiceStatus: contestcontracts.AWDServiceStatusUp, SLAScore: 1, DefenseScore: 2, AttackScore: 0, CheckerType: contestcontracts.AWDCheckerTypeHTTPStandard},
				},
				Attacks: []awdAttackSeed{
					{Offset: 5 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, SubmittedFlag: "awd{seed-test}", Source: contestcontracts.AWDAttackSourceSubmission, SubmittedByStudent: true, IsSuccess: true, ScoreGained: 30},
				},
				Traffic: []awdTrafficSeed{
					{Offset: 3 * time.Minute, AttackerTeamIndex: 0, VictimTeamIndex: 1, Method: "GET", Path: "/health", StatusCode: 200, Source: contestcontracts.AWDTrafficSourceRuntimeProxy},
				},
			},
		},
	}

	base := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	if err := seedStudentAWDScenario(db, teacher, student, scenario, catalog, base); err != nil {
		t.Fatalf("seed awd scenario: %v", err)
	}

	var service contestcontracts.ContestAWDService
	if err := db.First(&service).Error; err != nil {
		t.Fatalf("load contest awd service: %v", err)
	}
	if service.ID == 0 {
		t.Fatalf("expected generated service id")
	}
	if service.AWDChallengeID != 16 {
		t.Fatalf("unexpected awd challenge id: %d", service.AWDChallengeID)
	}

	var teamServices []contestcontracts.AWDTeamService
	if err := db.Order("id asc").Find(&teamServices).Error; err != nil {
		t.Fatalf("load awd team services: %v", err)
	}
	if len(teamServices) != 2 {
		t.Fatalf("expected 2 awd team services, got %d", len(teamServices))
	}
	for _, row := range teamServices {
		if row.ServiceID != service.ID {
			t.Fatalf("team service uses orphan service id %d, want %d", row.ServiceID, service.ID)
		}
	}

	var attack contestcontracts.AWDAttackLog
	if err := db.First(&attack).Error; err != nil {
		t.Fatalf("load awd attack log: %v", err)
	}
	if attack.ServiceID != service.ID {
		t.Fatalf("attack log uses orphan service id %d, want %d", attack.ServiceID, service.ID)
	}

	var traffic contestcontracts.AWDTrafficEvent
	if err := db.First(&traffic).Error; err != nil {
		t.Fatalf("load awd traffic event: %v", err)
	}
	if traffic.ServiceID != service.ID {
		t.Fatalf("traffic event uses orphan service id %d, want %d", traffic.ServiceID, service.ID)
	}
}

func newSyntheticChallengeCatalog(t *testing.T, perCategory int) *challengeCatalog {
	t.Helper()
	if perCategory <= 0 {
		t.Fatal("perCategory must be positive")
	}

	difficulties := []string{
		challengeentity.ChallengeDifficultyBeginner,
		challengeentity.ChallengeDifficultyEasy,
		challengeentity.ChallengeDifficultyMedium,
		challengeentity.ChallengeDifficultyHard,
	}
	catalog := &challengeCatalog{byCategory: make(map[string][]challengeRef, len(taxonomy.AllDimensions))}
	var nextID int64 = 1
	for _, dimension := range taxonomy.AllDimensions {
		items := make([]challengeRef, 0, perCategory)
		for index := 0; index < perCategory; index++ {
			items = append(items, challengeRef{
				ID:         nextID,
				Title:      fmt.Sprintf("%s-%02d", dimension, index+1),
				Category:   dimension,
				Difficulty: difficulties[index%len(difficulties)],
				Points:     100 + index*10,
				FlagType:   challengeentity.FlagTypeStatic,
			})
			nextID++
		}
		catalog.byCategory[dimension] = items
	}
	return catalog
}

func newSparseSyntheticChallengeCatalog(t *testing.T, perCategory map[string]int) *challengeCatalog {
	t.Helper()
	if len(perCategory) == 0 {
		t.Fatal("perCategory must not be empty")
	}

	difficulties := []string{
		challengeentity.ChallengeDifficultyBeginner,
		challengeentity.ChallengeDifficultyEasy,
		challengeentity.ChallengeDifficultyMedium,
		challengeentity.ChallengeDifficultyHard,
	}
	catalog := &challengeCatalog{byCategory: make(map[string][]challengeRef, len(taxonomy.AllDimensions))}
	var nextID int64 = 1
	for _, dimension := range taxonomy.AllDimensions {
		count := perCategory[dimension]
		if count <= 0 {
			continue
		}
		items := make([]challengeRef, 0, count)
		for index := 0; index < count; index++ {
			items = append(items, challengeRef{
				ID:         nextID,
				Title:      fmt.Sprintf("%s-%02d", dimension, index+1),
				Category:   dimension,
				Difficulty: difficulties[index%len(difficulties)],
				Points:     100 + index*10,
				FlagType:   challengeentity.FlagTypeStatic,
			})
			nextID++
		}
		catalog.byCategory[dimension] = items
	}
	return catalog
}

func weakestScenarioDimension(scenario studentScenario) string {
	weakestDimension := ""
	weakestScore := 2.0
	for _, dimension := range taxonomy.AllDimensions {
		score, ok := scenario.Profiles[dimension]
		if !ok {
			continue
		}
		if score < weakestScore {
			weakestScore = score
			weakestDimension = dimension
		}
	}
	return weakestDimension
}

func syntheticTopRecommendationChallengeID(
	t *testing.T,
	catalog *challengeCatalog,
	scenario studentScenario,
	dimension string,
) int64 {
	t.Helper()

	items := catalog.byCategory[dimension]
	if len(items) == 0 {
		t.Fatalf("missing catalog items for dimension %s", dimension)
	}

	solvedByID := make(map[int64]struct{}, len(scenario.Sessions))
	for _, session := range scenario.Sessions {
		challenge, err := catalog.pick(session.ChallengeCategory, session.ChallengeIndex)
		if err != nil {
			t.Fatalf("pick challenge for %s[%d]: %v", session.ChallengeCategory, session.ChallengeIndex, err)
		}
		for _, submission := range session.Submissions {
			if !submission.Correct {
				continue
			}
			solvedByID[challenge.ID] = struct{}{}
			break
		}
	}

	for _, item := range items {
		if _, solved := solvedByID[item.ID]; solved {
			continue
		}
		return item.ID
	}
	t.Fatalf("expected unsolved recommendation candidate for dimension %s", dimension)
	return 0
}

func syntheticMaxWrongStreak(scenario studentScenario) int {
	type submissionEvent struct {
		at      time.Duration
		correct bool
	}

	events := make([]submissionEvent, 0, len(scenario.Sessions)*3)
	for _, session := range scenario.Sessions {
		for _, submission := range session.Submissions {
			events = append(events, submissionEvent{
				at:      session.StartOffset + submission.Offset,
				correct: submission.Correct,
			})
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].at < events[j].at
	})

	maxStreak := 0
	current := 0
	for _, event := range events {
		if event.correct {
			current = 0
			continue
		}
		current++
		if current > maxStreak {
			maxStreak = current
		}
	}
	return maxStreak
}

func createSeedReviewUser(t *testing.T, db *gorm.DB, id int64, username, name, role string, now time.Time) *identitycontracts.User {
	t.Helper()

	user := &identitycontracts.User{
		ID:        id,
		Username:  username,
		Name:      name,
		Email:     fmt.Sprintf("%s@example.edu.cn", username),
		Role:      role,
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	switch role {
	case identitycontracts.RoleTeacher:
		user.TeacherNo = fmt.Sprintf("T%d", id)
	default:
		user.ClassName = seedClassName
		user.StudentNo = fmt.Sprintf("S%d", id)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}
