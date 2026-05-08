Webhook Inspector 会帮运维同学抓取第三方 webhook 文档并做预览，方便他们在上线前检查字段结构。

开发同学做了一个“本地地址黑名单”，试图阻止它去访问宿主或内网接口，但实现非常粗糙。请像真实 AWD 一样分析这套业务代码，利用漏洞获取当前实例中的动态 Flag。

## 目标

1. 分析首页、示例 manifest 与预览逻辑。
2. 利用预览入口访问内部调试接口。
3. 读取当前实例中的动态 Flag。

## 访问入口

- `GET /`
- `GET /manifest/demo`
- `GET /preview?url=<value>`

## 补充说明

- checker 通过 `PUT /api/flag` 和 `GET /api/flag` 管理动态 Flag。
- 防守时只修改业务工作区代码，不要破坏 `/health` 和 `/api/flag` 运行契约。
