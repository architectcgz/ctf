package dto

type AWDServiceTemplateQuery struct {
	Keyword     string `form:"keyword"`
	ServiceType string `form:"service_type"`
	Status      string `form:"status"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	Size        int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
