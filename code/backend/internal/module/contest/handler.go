package contest

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateContest(c *gin.Context) {
	var req dto.CreateContestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateContest(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) UpdateContest(c *gin.Context) {
	id := c.GetInt64("id")
	var req dto.UpdateContestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.UpdateContest(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) GetContest(c *gin.Context) {
	id := c.GetInt64("id")
	resp, err := h.service.GetContest(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ListContests(c *gin.Context) {
	var req dto.ListContestsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	contests, total, err := h.service.ListContests(c.Request.Context(), &req)
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

	response.Page(c, contests, total, page, size)
}
