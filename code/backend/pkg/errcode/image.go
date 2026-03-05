package errcode

var (
	ErrImageAlreadyExists = New(12101, "镜像已存在", 409)
	ErrImageNotAccessible = New(12102, "Docker 镜像不存在或无法访问", 400)
	ErrImageNotFound      = New(12103, "镜像不存在", 404)
	ErrImageInUse         = New(12104, "镜像正在使用中，无法删除", 400)
)
