package application

import "context"

// ContainerStatsService 收口受管容器运行指标查询能力。
type ContainerStatsService struct {
	reader ManagedContainerStatsReader
}

// NewContainerStatsService 创建受管容器指标查询服务。
func NewContainerStatsService(reader ManagedContainerStatsReader) *ContainerStatsService {
	return &ContainerStatsService{reader: reader}
}

// ListManagedContainerStats 返回受管容器指标快照。
func (s *ContainerStatsService) ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error) {
	if s == nil || s.reader == nil {
		return []ManagedContainerStat{}, nil
	}
	return s.reader.ListManagedContainerStats(normalizeContext(ctx))
}
