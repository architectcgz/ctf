package queries

import (
	"context"

	runtimeapp "ctf-platform/internal/module/container_runtime/application"
)

// ContainerStatsService 收口受管容器运行指标查询能力。
type ContainerStatsService struct {
	reader runtimeapp.ManagedContainerStatsReader
}

// NewContainerStatsService 创建受管容器指标查询服务。
func NewContainerStatsService(reader runtimeapp.ManagedContainerStatsReader) *ContainerStatsService {
	return &ContainerStatsService{reader: reader}
}

// ListManagedContainerStats 返回受管容器指标快照。
func (s *ContainerStatsService) ListManagedContainerStats(ctx context.Context) ([]runtimeapp.ManagedContainerStat, error) {
	if s == nil || s.reader == nil {
		return []runtimeapp.ManagedContainerStat{}, nil
	}
	return s.reader.ListManagedContainerStats(ctx)
}
