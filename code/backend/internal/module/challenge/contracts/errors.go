package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var (
	ErrImageAlreadyExists = apperror.Define(12101, "镜像已存在", http.StatusConflict)
	ErrImageNotAccessible = apperror.Define(12102, "Docker 镜像不存在或无法访问", http.StatusBadRequest)
	ErrImageNotFound      = apperror.Define(12103, "镜像不存在", http.StatusNotFound)
	ErrImageInUse         = apperror.Define(12104, "镜像正在使用中，无法删除", http.StatusConflict)
)

var (
	ErrAlreadySolved       = apperror.Define(13002, "该题目已完成", http.StatusConflict)
	ErrSubmitTooFrequent   = apperror.Define(13003, "提交过于频繁，请稍后再试", http.StatusTooManyRequests)
	ErrChallengeNotFound   = apperror.Define(13004, "靶场不存在", http.StatusNotFound)
	ErrChallengeNotPublish = apperror.Define(
		13005,
		"靶场未发布",
		http.StatusForbidden,
	)
)
