package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) CreateRound(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateAWDRoundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	input := createAWDRoundInputFromDTO(&req)

	readinessSnapshot, err := loadAWDReadinessAuditSnapshot(c.Request.Context(), h.queries, contestID, input.ForceOverride)
	if err != nil {
		response.FromError(c, err)
		return
	}

	requestCtx, gateTrace := prepareAWDReadinessGateTrace(c.Request.Context(), readinessSnapshot)
	resp, err := h.commands.CreateRound(requestCtx, contestID, input)
	writeAWDReadinessAuditPayload(c, contestdomain.AWDReadinessActionCreateRound, input.OverrideReason, readinessSnapshot, gateTrace, err)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func awdRoundResultsToDTO(results []contestqry.AWDRoundResult) []*dto.AWDRoundResp {
	resp := make([]*dto.AWDRoundResp, 0, len(results))
	for i := range results {
		round := results[i]
		resp = append(resp, &dto.AWDRoundResp{
			ID:           round.ID,
			ContestID:    round.ContestID,
			RoundNumber:  round.RoundNumber,
			Status:       round.Status,
			StartedAt:    round.StartedAt,
			EndedAt:      round.EndedAt,
			AttackScore:  round.AttackScore,
			DefenseScore: round.DefenseScore,
			CreatedAt:    round.CreatedAt,
			UpdatedAt:    round.UpdatedAt,
		})
	}
	return resp
}

func (h *AWDHandler) ListRounds(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.queries.ListRounds(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdRoundResultsToDTO(resp))
}
