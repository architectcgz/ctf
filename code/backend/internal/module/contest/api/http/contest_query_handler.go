package http

import (
	"fmt"
	"strings"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetContest(c *gin.Context) {
	id := c.GetInt64("id")
	resp, err := h.queries.GetContest(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestRequestMapper.ToContestRespPtr(resp))
}

func (h *Handler) ListContests(c *gin.Context) {
	var req dto.ListContestsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	statuses, err := parseContestStatuses(req.Status, req.Statuses)
	if err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	query := contestqry.ListContestsInput{
		Statuses:  statuses,
		Mode:      req.Mode,
		SortKey:   req.SortKey,
		SortOrder: req.SortOrder,
		Page:      req.Page,
		Size:      req.Size,
	}

	contests, total, err := h.queries.ListContests(c.Request.Context(), query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	summary, err := h.queries.GetContestListSummary(c.Request.Context(), query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}

	response.Success(c, dto.ContestPageResp{
		List:     contestRequestMapper.ToContestResps(contests),
		Total:    total,
		Page:     page,
		PageSize: size,
		Summary: dto.ContestListSummaryResp{
			DraftCount:       summary.DraftCount,
			RegisteringCount: summary.RegistrationCount,
			RunningCount:     summary.RunningCount,
			FrozenCount:      summary.FrozenCount,
			EndedCount:       summary.EndedCount,
		},
	})
}

var allowedContestStatuses = map[string]struct{}{
	"draft":        {},
	"registration": {},
	"running":      {},
	"frozen":       {},
	"ended":        {},
}

func parseContestStatuses(status *string, csv string) ([]string, error) {
	values := make([]string, 0, 4)
	if status != nil && strings.TrimSpace(*status) != "" {
		values = append(values, strings.TrimSpace(*status))
	}
	if strings.TrimSpace(csv) != "" {
		for _, item := range strings.Split(csv, ",") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			values = append(values, trimmed)
		}
	}
	if len(values) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(values))
	statuses := make([]string, 0, len(values))
	for _, item := range values {
		if _, ok := allowedContestStatuses[item]; !ok {
			return nil, fmt.Errorf("无效的赛事状态筛选: %s", item)
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		statuses = append(statuses, item)
	}
	return statuses, nil
}
