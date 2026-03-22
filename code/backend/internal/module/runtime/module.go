package runtime

type Module struct {
	*Service
	proxyTickets         *ProxyTicketService
	proxyBodyPreviewSize int
}

func NewModule(service *Service, proxyTickets *ProxyTicketService, proxyBodyPreviewSize int) *Module {
	return &Module{
		Service:              service,
		proxyTickets:         proxyTickets,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
	}
}

func (m *Module) proxyTicketMaxAge() int {
	if m == nil || m.proxyTickets == nil || m.proxyTickets.cfg == nil {
		return 0
	}
	return int(m.proxyTickets.cfg.ProxyTicketTTL.Seconds())
}
