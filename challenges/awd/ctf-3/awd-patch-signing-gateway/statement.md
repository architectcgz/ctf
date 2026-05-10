Patch Signing Gateway 把公网入口、内部 signer 和 key vault 拆成了三段：

- `patch-gateway` 对外提供补丁 bundle 入口
- `signer-core` 负责内部签名
- `key-vault` 保存签名元数据

开发同学为了兼容不同岗位，把 signer 权限角色交给公网入口代填。请结合这三段服务的信任关系，像真实 AWD 一样分析业务行为，获取当前实例中的动态 Flag。

## 目标

1. 分析首页、示例渠道与 bundle 逻辑。
2. 找到公网入口如何间接影响内部 signer / vault 链路。
3. 读取当前实例中的动态 Flag。

## 访问入口

- `GET /`
- `GET /channels/demo`
- `GET /bundle?channel=<value>&signer_role=<value>`

## 补充说明

- checker 通过 `PUT /api/flag` 和 `GET /api/flag` 管理动态 Flag。
- 防守时只修改业务工作区代码，不要破坏 `/health` 和 `/api/flag` 运行契约。
- 这题的关键在于客户端参数是如何变成内部 signer 的权限角色。
