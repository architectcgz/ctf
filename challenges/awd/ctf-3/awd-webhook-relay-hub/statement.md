Webhook Relay Hub 把公网入口、内部回放节点和归档节点拆成了三段：

- `relay-web` 对外提供 feed 预览
- `relay-worker` 负责内部 replay
- `relay-store` 保存归档回放数据

开发同学为了图省事，只在公网预览器里做了很粗糙的地址黑名单。请结合这三段服务的协作关系，像真实 AWD 一样分析业务行为，获取当前实例中的动态 Flag。

## 目标

1. 分析首页、示例 feed 与预览逻辑。
2. 找到公网节点如何间接访问内部 replay / archive 链路。
3. 读取当前实例中的动态 Flag。

## 访问入口

- `GET /`
- `GET /feeds/demo`
- `GET /preview?url=<value>`

## 补充说明

- checker 通过 `PUT /api/flag` 和 `GET /api/flag` 管理动态 Flag。
- 防守时只修改业务工作区代码，不要破坏 `/health` 和 `/api/flag` 运行契约。
- 这题的关键不在单个 Web 漏洞，而在三段服务之间的访问边界。
