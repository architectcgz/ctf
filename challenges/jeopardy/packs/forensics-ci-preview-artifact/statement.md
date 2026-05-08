团队把一份 CI 预发布工件包误放到了公开预览页，里面包含运行日志和步骤摘要。

页面还保留了一个需要 token 的导出接口。请分析工件包里的线索，恢复这个 token，并拿到实例中的动态 Flag。

## 目标

1. 下载 `preview-bundle.zip`。
2. 分析日志和摘要材料，恢复 token。
3. 调用导出接口拿到动态 Flag。

## 获取方式

- `/` 首页
- `/download/preview-bundle.zip`
- `/redeem?token=<value>`

## 补充说明

- 工件包来自一次真实预发布流程，线索散落在多个文本文件中。
- 不需要爆破 token。
