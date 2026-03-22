package teacher

import teachinghttp "ctf-platform/internal/module/teaching_readmodel/api/http"

type Handler = teachinghttp.Handler

func NewHandler(service *Service) *Handler {
	return teachinghttp.NewHandler(service)
}
