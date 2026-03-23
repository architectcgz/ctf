# CTF Backend 代码风格补充约束

## 1. 文档定位

CTF 后端默认先遵守 workspace 共享 Go 规范：

- `/home/azhi/workspace/docs/go-code-style-guide.md`

这份文档只补充 CTF 自己的项目化约束，重点回答三个问题：

- 共享规范里的 owner 在 CTF 中如何映射
- CTF 模块继续迁移时应收敛到什么结构
- 哪些历史写法在 CTF 中必须继续清理

## 2. CTF 的 owner 映射规则

共享规范里的：

- `service boundary`
- `owner`

在 CTF 中统一映射成：

- `module boundary`

也就是说，CTF 的最终目标不是多服务单仓，而是：

- 单体部署
- 模块化单体
- 模块内部严格分层
- 模块之间通过窄 contracts / query facade / event 交互

## 3. CTF 后端默认模块结构

新增模块或较大重构时，优先采用：

```text
internal/module/<module>/
├── contracts.go
├── api/http/
├── application/commands/
├── application/queries/
├── domain/
├── ports/
├── infrastructure/<adapter>/
└── runtime/
```

说明：

- `contracts.go`：模块对外暴露的最小能力面
- `api/http`：Gin handler，仅做协议映射
- `application/commands`：写用例、事务边界、状态推进
- `application/queries`：读用例、只读聚合、投影视图拼装
- `domain`：实体、规则、状态机、领域错误
- `ports`：由消费方定义的最小依赖接口
- `infrastructure/<adapter>`：GORM、Redis、Docker、导出、外部集成等适配器
- `runtime/`：模块内唯一 wiring 层；如果模块当前没有装配需求，可以暂时不建

## 4. CTF 的额外硬约束

### 4.1 不再新增兼容层

CTF 当前迁移采用硬迁移，不保留长期兼容层。

因此：

- 不再新增根包 `module.go` 门面层
- 不再新增旧接口到新实现的转发壳
- 不再为了“先跑起来”引入新的桥接 module

如果某次迁移不得不出现临时转发文件，它也必须被视为待删除遗留，并写进对应 `.planning/` 任务，而不是进入目标结构。

### 4.2 composition 是唯一全局装配入口

全局装配统一收敛到：

- `code/backend/internal/app/composition`

模块内部如需局部装配，统一放在本模块的 `runtime/`。
禁止在 handler、repository、测试辅助包里偷偷拼跨模块依赖。

### 4.3 跨模块调用必须收窄

跨模块调用优先采用：

- `contracts.go`
- query facade
- 领域事件

不允许直接依赖对方的：

- concrete repository
- concrete service
- `infrastructure`
- `api/http`

### 4.4 readmodel 是正式 owner，不是临时查询包

满足以下任一条件时，应优先拆出 readmodel：

- 查询会聚合多个 owner 的数据
- 读路径和写路径关注点明显不同
- 查询逻辑已经开始超过命令模块的可维护范围
- 需要单独治理分页、投影、时间线、统计或推荐视图

已经成立的方向包括：

- `practice_readmodel`
- `teaching_readmodel`

后续迁移时，应继续防止读逻辑回流到命令模块。

### 4.5 错误处理统一收敛到项目既有体系

CTF 的 canonical 错误表达继续使用项目现有错误码体系收敛。

约束：

- `infrastructure` 返回原始错误或带上下文包装的错误
- `application` 负责收敛到项目错误码
- `api/http` 负责映射到响应结构
- 不要跨层依赖裸字符串错误做业务判断

### 4.6 缓存、配置与副作用必须显式

- cache key 统一收敛到明确常量或 adapter
- TTL 必须从配置读取，不写死在业务逻辑里
- Docker、导出、审计、异步任务等副作用只能出现在应用层编排点或 adapter 中

## 5. 迁移时的 review 检查项

每次后端迁移默认检查：

1. handler 是否仍在编排业务或直接调 repository
2. `application` 是否直接依赖 concrete adapter
3. 是否跨模块 import 了对方内部实现
4. 是否把读侧聚合又塞回了命令模块
5. 是否继续制造根包兼容壳或新的桥接层
6. cache、配置、事务边界是否显式

## 6. 增量迁移要求

这份补充约束不是要求一次性重写全仓，而是要求：

1. 新代码按共享规范和本地补充一起落地
2. 当前正在迁移的模块优先向目标结构物理收敛
3. 旧模块在被持续修改时逐步补齐 `domain / ports / runtime`
4. 不再继续制造新的大平铺文件和历史兼容层
