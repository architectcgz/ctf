package challenge

import "errors"

var (
	ErrImageNotFound         = errors.New("镜像不存在")
	ErrHasRunningInstances   = errors.New("存在运行中的实例，无法删除")
	ErrImageNotLinked        = errors.New("靶场未关联镜像，无法发布")
	ErrChallengeNotPublished = errors.New("靶场未发布")
)
