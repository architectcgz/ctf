package http

import opsports "ctf-platform/internal/module/ops/ports"

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type dashboardResponseMapperContract interface {
	ToDashboardStats(source opsports.DashboardStatsSnapshot) DashboardStats
	ToDashboardStatsPtr(source *opsports.DashboardStatsSnapshot) *DashboardStats
	ToContainerStat(source opsports.DashboardContainerStat) ContainerStat
	ToResourceAlert(source opsports.DashboardResourceAlert) ResourceAlert
}

var dashboardResponseMapper dashboardResponseMapperContract
