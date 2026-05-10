Preview Render Farm 把公网入口、内部 renderer 和素材缓存拆成了三段：

- `render-web` 对外提供预览入口
- `render-worker` 负责内部渲染
- `asset-cache` 保存素材与调试缓存

开发同学为了追求“快速预览”，把素材路径直接拼进了内部渲染链路。请结合这三段服务的协作关系，像真实 AWD 一样分析业务行为，获取当前实例中的动态 Flag。

## 目标

1. 分析首页、素材目录与预览逻辑。
2. 找到公网节点如何间接访问内部 renderer / cache 链路。
3. 读取当前实例中的动态 Flag。

## 访问入口

- `GET /`
- `GET /catalog/demo`
- `GET /preview?asset=<value>`

## 补充说明

- checker 通过 `PUT /api/flag` 和 `GET /api/flag` 管理动态 Flag。
- 防守时只修改业务工作区代码，不要破坏 `/health` 和 `/api/flag` 运行契约。
- 这题的关键在于素材路径如何被一路带进内部渲染流水线。
