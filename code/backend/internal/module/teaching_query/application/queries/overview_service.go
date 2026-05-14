package queries

import (
	"context"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
	"ctf-platform/pkg/errcode"
)

type overviewQueryRepository interface {
	queryports.TeachingClassQueryRepository
	queryports.TeachingStudentDirectoryRepository
	queryports.TeachingClassInsightRepository
	queryports.TeachingOverviewRepository
}

type OverviewQueryService struct {
	users queryports.TeachingUserLookupRepository
	repo  overviewQueryRepository
}

var _ OverviewService = (*OverviewQueryService)(nil)

func NewOverviewService(
	users queryports.TeachingUserLookupRepository,
	repo overviewQueryRepository,
) *OverviewQueryService {
	return &OverviewQueryService{
		users: users,
		repo:  repo,
	}
}

func (s *OverviewQueryService) GetOverview(ctx context.Context, requesterID int64, requesterRole string) (*dto.TeacherOverviewResp, error) {
	classItems, err := s.listAccessibleClassItems(ctx, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}
	if len(classItems) == 0 {
		return emptyOverviewResponse(), nil
	}

	classNames := make([]string, 0, len(classItems))
	for _, item := range classItems {
		if name := strings.TrimSpace(item.Name); name != "" {
			classNames = append(classNames, name)
		}
	}
	if len(classNames) == 0 {
		return emptyOverviewResponse(), nil
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())

	studentItems, err := s.repo.ListStudentsByClasses(ctx, classNames, "", "", startOfDay)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	students := commonmapper.NonNilSlice(teachingQueryMapper.ToStudentItems(studentItems))

	trend, err := s.repo.GetOverviewTrend(ctx, classNames, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	focusClasses, err := s.buildOverviewClassFocuses(ctx, classItems, startOfDay)
	if err != nil {
		return nil, err
	}

	focusStudents := selectOverviewRiskStudents(students, 6)
	summary := buildOverviewSummary(classItems, students, focusStudents)
	spotlightStudent := selectOverviewTopStudent(students)

	return &dto.TeacherOverviewResp{
		Summary:          summary,
		Trend:            mapOverviewTrend(trend),
		FocusClasses:     focusClasses,
		FocusStudents:    focusStudents,
		SpotlightStudent: spotlightStudent,
		WeakDimensions:   buildOverviewWeakDimensions(students),
	}, nil
}

func emptyOverviewResponse() *dto.TeacherOverviewResp {
	return &dto.TeacherOverviewResp{
		Summary:        dto.TeacherOverviewSummaryResp{},
		Trend:          dto.TeacherOverviewTrendResp{Points: []dto.TeacherOverviewTrendPoint{}},
		FocusClasses:   []dto.TeacherOverviewClassFocusResp{},
		FocusStudents:  []dto.TeacherStudentItem{},
		WeakDimensions: []dto.TeacherOverviewWeakDimensionResp{},
	}
}

func (s *OverviewQueryService) listAccessibleClassItems(ctx context.Context, requesterID int64, requesterRole string) ([]queryports.ClassItem, error) {
	if requesterRole == model.RoleAdmin {
		items, err := s.repo.ListClasses(ctx, 0, 0)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return items, nil
	}

	requester, err := s.users.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return nil, errcode.ErrUnauthorized
	}

	className := strings.TrimSpace(requester.ClassName)
	if className == "" {
		return []queryports.ClassItem{}, nil
	}

	count, err := s.repo.CountStudentsByClass(ctx, className)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return []queryports.ClassItem{{
		Name:         className,
		StudentCount: count,
	}}, nil
}

func selectOverviewRiskStudents(students []dto.TeacherStudentItem, limit int) []dto.TeacherStudentItem {
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

	return limitOverviewStudents(filtered, limit)
}

func selectOverviewTopStudent(students []dto.TeacherStudentItem) *dto.TeacherStudentItem {
	if len(students) == 0 {
		return nil
	}

	sorted := append([]dto.TeacherStudentItem(nil), students...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].SolvedCount != sorted[j].SolvedCount {
			return sorted[i].SolvedCount > sorted[j].SolvedCount
		}
		if sorted[i].TotalScore != sorted[j].TotalScore {
			return sorted[i].TotalScore > sorted[j].TotalScore
		}
		return sorted[i].Username < sorted[j].Username
	})

	top := sorted[0]
	return &top
}

func selectOverviewWeakDimensionStudents(students []dto.TeacherStudentItem) (string, []dto.TeacherStudentItem) {
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

func limitOverviewStudents(students []dto.TeacherStudentItem, limit int) []dto.TeacherStudentItem {
	if limit <= 0 || len(students) <= limit {
		return students
	}
	return students[:limit]
}

func (s *OverviewQueryService) buildOverviewClassFocuses(
	ctx context.Context,
	classItems []queryports.ClassItem,
	since time.Time,
) ([]dto.TeacherOverviewClassFocusResp, error) {
	focuses := make([]dto.TeacherOverviewClassFocusResp, 0, len(classItems))
	for _, item := range classItems {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}

		summary, err := s.repo.GetClassSummary(ctx, item.Name, since)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		studentItems, err := s.repo.ListStudentsByClass(ctx, item.Name, "", "", since)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		students := commonmapper.NonNilSlice(teachingQueryMapper.ToStudentItems(studentItems))
		dominantWeakDimension, _ := selectOverviewWeakDimensionStudents(students)
		riskStudents := selectOverviewRiskStudents(students, len(students))

		focuses = append(focuses, dto.TeacherOverviewClassFocusResp{
			ClassName:             item.Name,
			StudentCount:          summary.StudentCount,
			ActiveRate:            summary.ActiveRate,
			RecentEventCount:      summary.RecentEventCount,
			RiskStudentCount:      int64(len(riskStudents)),
			DominantWeakDimension: dominantWeakDimension,
		})
	}

	sort.Slice(focuses, func(i, j int) bool {
		if focuses[i].RiskStudentCount != focuses[j].RiskStudentCount {
			return focuses[i].RiskStudentCount > focuses[j].RiskStudentCount
		}
		if focuses[i].RecentEventCount != focuses[j].RecentEventCount {
			return focuses[i].RecentEventCount > focuses[j].RecentEventCount
		}
		return focuses[i].ClassName < focuses[j].ClassName
	})

	if len(focuses) > 6 {
		return focuses[:6], nil
	}
	return focuses, nil
}

func buildOverviewSummary(
	classItems []queryports.ClassItem,
	students []dto.TeacherStudentItem,
	focusStudents []dto.TeacherStudentItem,
) dto.TeacherOverviewSummaryResp {
	summary := dto.TeacherOverviewSummaryResp{
		ClassCount:       int64(len(classItems)),
		StudentCount:     int64(len(students)),
		RiskStudentCount: int64(len(focusStudents)),
	}
	if len(students) == 0 {
		return summary
	}

	totalSolved := 0
	for _, student := range students {
		totalSolved += student.SolvedCount
		summary.RecentEventCount += int64(student.RecentEventCount)
		if student.RecentEventCount > 0 {
			summary.ActiveStudentCount++
		}
	}

	summary.AverageSolved = float64(totalSolved) / float64(len(students))
	summary.ActiveRate = float64(summary.ActiveStudentCount) * 100 / float64(len(students))
	return summary
}

func mapOverviewTrend(source *queryports.OverviewTrend) dto.TeacherOverviewTrendResp {
	if source == nil || len(source.Points) == 0 {
		return dto.TeacherOverviewTrendResp{Points: []dto.TeacherOverviewTrendPoint{}}
	}

	points := make([]dto.TeacherOverviewTrendPoint, 0, len(source.Points))
	for _, point := range source.Points {
		points = append(points, dto.TeacherOverviewTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}
	return dto.TeacherOverviewTrendResp{Points: points}
}

func buildOverviewWeakDimensions(students []dto.TeacherStudentItem) []dto.TeacherOverviewWeakDimensionResp {
	counter := make(map[string]int64)
	for _, student := range students {
		if student.WeakDimension == nil {
			continue
		}
		key := strings.TrimSpace(*student.WeakDimension)
		if key == "" {
			continue
		}
		counter[key]++
	}
	if len(counter) == 0 {
		return []dto.TeacherOverviewWeakDimensionResp{}
	}

	items := make([]dto.TeacherOverviewWeakDimensionResp, 0, len(counter))
	for dimension, count := range counter {
		items = append(items, dto.TeacherOverviewWeakDimensionResp{
			Dimension:    dimension,
			StudentCount: count,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].StudentCount != items[j].StudentCount {
			return items[i].StudentCount > items[j].StudentCount
		}
		return items[i].Dimension < items[j].Dimension
	})
	if len(items) > 6 {
		return items[:6]
	}
	return items
}
