package contest

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChallengeHandler struct {
	service *ChallengeService
}

func NewChallengeHandler(service *ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{service: service}
}

func (h *ChallengeHandler) AddChallenge(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.AddContestChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrValidationFailed)
		return
	}

	resp, err := h.service.AddChallengeToContest(contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ChallengeHandler) RemoveChallenge(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("challenge_id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	if err := h.service.RemoveChallengeFromContest(contestID, challengeID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ChallengeHandler) UpdatePoints(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("challenge_id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.UpdateContestChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrValidationFailed)
		return
	}

	if err := h.service.UpdateChallengePoints(contestID, challengeID, &req); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	challenges, err := h.service.GetContestChallenges(contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, challenges)
}
