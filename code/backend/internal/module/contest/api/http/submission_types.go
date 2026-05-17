package http

type SubmitFlagReq struct {
	Flag string `json:"flag" binding:"required"`
}
