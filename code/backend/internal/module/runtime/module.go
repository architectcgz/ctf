package runtime

import "ctf-platform/internal/module/container"

type Module struct {
	*container.Service
}

func NewModule(service *container.Service) *Module {
	return &Module{Service: service}
}
