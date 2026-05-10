Forwarded Admin Gateway 把公网入口、内部管理面和审计数据面拆成了三段：

- `edge-gateway` 对外提供转发入口
- `admin-app` 负责内部管理导出
- `audit-db` 保存租户审计数据

开发同学为了方便排查，把一部分管理面能力挂到了公网网关后面。请结合这三段服务的信任关系，像真实 AWD 一样分析业务行为，获取当前实例中的动态 Flag。

## 目标

1. 分析首页、示例报告与公网转发逻辑。
2. 找到公网网关如何间接打到内部 admin / audit 链路。
3. 读取当前实例中的动态 Flag。

## 访问入口

- `GET /`
- `GET /reports/demo`
- `GET /proxy?path=<value>`

## 补充说明

- checker 通过 `PUT /api/flag` 和 `GET /api/flag` 管理动态 Flag。
- 防守时只修改业务工作区代码，不要破坏 `/health` 和 `/api/flag` 运行契约。
- 这题的关键不在单个表单输入，而在网关如何把用户头部透传给内部服务。
