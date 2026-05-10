IoT Fleet Orchestrator 把公网控制台、内部 agent 和配置仓库拆成了三段：

- `fleet-web` 对外提供控制台入口
- `fleet-agent` 负责内部设备调度
- `config-vault` 保存设备配置

开发同学为了方便运维留了一个支持默认 key。请结合这三段服务的信任关系，像真实 AWD 一样分析业务行为，获取当前实例中的动态 Flag。

## 目标

1. 分析首页、示例设备与公网 dispatch 逻辑。
2. 找到公网控制台如何间接访问内部 agent / vault 链路。
3. 读取当前实例中的动态 Flag。

## 访问入口

- `GET /`
- `GET /devices/demo`
- `GET /dispatch?device=<id>&key=<value>&op=<value>`

## 补充说明

- checker 通过 `PUT /api/flag` 和 `GET /api/flag` 管理动态 Flag。
- 防守时只修改业务工作区代码，不要破坏 `/health` 和 `/api/flag` 运行契约。
- 这题的关键在于默认 key 如何穿透到内部设备控制链路。
