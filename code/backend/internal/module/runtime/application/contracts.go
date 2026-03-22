package application

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

// CountRunningRepository 定义运行中实例统计仓储能力。
type CountRunningRepository interface {
	CountRunning() (int64, error)
}

// InstanceRepository 定义实例 HTTP 用例所需的仓储能力。
type InstanceRepository interface {
	FindByID(id int64) (*model.Instance, error)
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*model.Instance, error)
	ListVisibleByUser(ctx context.Context, userID int64) ([]UserVisibleInstanceRow, error)
	ListTeacherInstances(ctx context.Context, filter TeacherInstanceFilter) ([]TeacherInstanceRow, error)
	AtomicExtendByIDWithContext(ctx context.Context, id int64, maxExtends int, duration time.Duration) error
	UpdateStatusAndReleasePort(id int64, status string) error
}

// RuntimeCleaner 定义实例销毁时的运行时资源清理能力。
type RuntimeCleaner interface {
	CleanupRuntimeWithContext(ctx context.Context, instance *model.Instance) error
}

// ManagedContainerStat 表示 runtime application 层暴露的受管容器运行指标快照。
type ManagedContainerStat struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   int64
	MemoryLimit   int64
}

// ManagedContainer 表示 runtime application 层暴露的受管容器元数据。
type ManagedContainer struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// TeacherInstanceFilter 定义教师端实例列表筛选条件。
type TeacherInstanceFilter struct {
	ClassName string
	Keyword   string
	StudentNo string
}

// UserVisibleInstanceRow 表示用户可见实例列表行模型。
type UserVisibleInstanceRow struct {
	ID             int64
	ChallengeID    int64
	ChallengeTitle string
	Category       string
	Difficulty     string
	FlagType       string
	Status         string
	AccessURL      string
	ExpiresAt      time.Time
	ExtendCount    int
	MaxExtends     int
	CreatedAt      time.Time
}

// TeacherInstanceRow 表示教师端实例列表行模型。
type TeacherInstanceRow struct {
	ID              int64
	StudentID       int64
	StudentName     string
	StudentUsername string
	StudentNo       *string
	ClassName       string
	ChallengeID     int64
	ChallengeTitle  string
	Status          string
	AccessURL       string
	ExpiresAt       time.Time
	ExtendCount     int
	MaxExtends      int
	CreatedAt       time.Time
}
