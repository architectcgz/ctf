# 数据库 Migration 管理

> 状态：Current
> 事实源：`code/backend/migrations/`、`code/backend/cmd/migrate/main.go`
> 替代：无

## 本文档范围

| 覆盖 | 不覆盖 |
|------|--------|
| Migration 编号规则与文件命名约定 | PostgreSQL 安装与配置 |
| 本地开发 migration 应用流程 | 生产部署的完整运维手册 |
| Migration 回滚策略与风险评估 | 数据库连接池配置（见 `database-tuning.md`） |
| 代码位置与工具入口 | 完整灾备方案 |

## 定位

本文档只说明后端数据库 migration 的增量更新机制、编号规则、本地应用流程和回滚策略。

## 当前设计

- `code/backend/migrations/`
  - 负责：存放所有数据库 schema 变更的 `.up.sql` 和 `.down.sql` 文件对
  - 不负责：运行时数据迁移、种子数据或测试 fixture

- `code/backend/cmd/migrate/main.go`
  - 负责：Migration 命令行工具入口，支持 `up`、`down`、`version` 等命令
  - 不负责：自动检测并应用 migration

## 1. Migration 编号规则

当前 migration 文件命名格式：

```
<6位递增序号>_<描述性名称>.<up|down>.sql
```

示例：

- `000015_remove_legacy_awd_runtime_config_challenge_id.up.sql`
- `000015_remove_legacy_awd_runtime_config_challenge_id.down.sql`
- `000018_create_platform_event_outbox.up.sql`
- `000018_create_platform_event_outbox.down.sql`

规则：

- 序号必须连续递增，从 `000001` 开始
- 每个 migration 必须同时提供 `.up.sql` 和 `.down.sql`
- 描述性名称使用小写字母和下划线，简洁说明变更内容
- 已合并到 `main` 的 migration 编号不得修改

## 2. 本地开发流程

### 2.1 应用 Migration

```bash
# 进入后端目录
cd code/backend

# 应用所有待执行的 migration
go run cmd/migrate/main.go up

# 应用指定数量的 migration
go run cmd/migrate/main.go up -n 1
```

### 2.2 回滚 Migration

```bash
# 回滚最近一次 migration
go run cmd/migrate/main.go down -n 1

# 回滚到指定版本
go run cmd/migrate/main.go version <版本号>
```

### 2.3 查看当前版本

```bash
go run cmd/migrate/main.go version
```

### 2.4 新增 Migration

1. 确定下一个可用序号（查看 `migrations/` 目录最大序号）
2. 创建 `.up.sql` 和 `.down.sql` 文件对
3. 在 `.up.sql` 中编写 schema 变更
4. 在 `.down.sql` 中编写对应回滚逻辑
5. 本地测试：先 `up` 再 `down` 验证可逆性

## 3. Migration 文件命名约定

描述性名称应清晰说明变更类型：

| 变更类型 | 命名示例 |
|---------|---------|
| 创建表 | `create_<table_name>` |
| 删除表 | `drop_<table_name>` |
| 添加列 | `add_<column_name>_to_<table_name>` |
| 删除列 | `remove_<column_name>_from_<table_name>` |
| 修改列 | `alter_<column_name>_in_<table_name>` |
| 添加索引 | `add_index_on_<table_name>_<column_name>` |
| 删除索引 | `drop_index_on_<table_name>_<column_name>` |
| 重命名 | `rename_<old_name>_to_<new_name>` |
| 数据清理 | `cleanup_<描述>` |

## 4. 生产环境 Migration 策略

### 4.1 风险评估

执行 migration 前必须评估：

- **锁表风险**：`ALTER TABLE` 是否会长时间锁表
- **数据量影响**：大表上的 schema 变更可能耗时数分钟
- **回滚可行性**：`.down.sql` 是否真的可以安全回滚
- **数据丢失风险**：`DROP COLUMN` 或 `DROP TABLE` 必须确认数据已迁移

### 4.2 回滚策略

**可安全回滚的场景**：

- 添加可空列
- 添加索引
- 创建新表（无数据）

**不可安全回滚的场景**：

- 删除列或表（数据已丢失）
- 修改列类型（可能有数据截断）
- 删除索引后已有大量写入（回建索引耗时长）

**回滚决策**：

- 优先考虑前滚修复（forward fix），而非回滚
- 回滚前必须确认 `.down.sql` 已测试
- 回滚前必须备份当前数据库状态

### 4.3 生产环境执行流程

1. **备份数据库**：`pg_dump` 或云平台快照
2. **在测试环境验证**：先在与生产一致的数据集上测试
3. **通知相关方**：告知预期停机时间或性能影响
4. **执行 migration**：使用生产部署工具链
5. **验证结果**：检查应用日志、数据完整性
6. **保留备份**：至少保留到下次成功部署

## 5. 常见问题

### 5.1 Migration 冲突

**场景**：两个分支同时添加了相同编号的 migration

**解决**：

1. 合并时重新编号其中一个 migration
2. 更新文件名中的序号
3. 本地重新执行 migration 测试

### 5.2 Migration 失败

**场景**：`up` 执行到一半失败

**解决**：

1. 检查错误日志，定位失败原因
2. 手动检查数据库状态（部分变更可能已生效）
3. 根据情况选择：
   - 修复数据后重新执行
   - 手动回滚部分变更后执行 `down`
   - 标记当前版本为失败，跳过该 migration

### 5.3 忘记写 `.down.sql`

**场景**：只提交了 `.up.sql`

**解决**：

- Migration 工具会拒绝执行
- 必须补充 `.down.sql`，即使只是占位符（不可逆 migration 应在注释中说明）

## 6. 时间列约定

新增时间列时遵循以下约定：

```sql
-- 使用 TIMESTAMPTZ，默认值为 now()
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

-- 可空时间列
last_seen_at TIMESTAMPTZ NULL,
deleted_at TIMESTAMPTZ NULL
```

说明：

- 统一使用 `TIMESTAMPTZ` 而非 `TIMESTAMP`
- 后端代码使用 `time.Now().UTC()` 与数据库保持一致
- PostgreSQL 连接必须显式声明 `TimeZone=UTC`

## 7. 边界

### 7.1 Migration 不承载的职责

- **种子数据**：测试数据应放在 `internal/testutil/` 或测试 fixture
- **数据迁移**：大规模数据转换应单独编写迁移脚本
- **环境配置**：数据库连接参数、权限配置不在 migration 中

### 7.2 与其他文档的关系

- 数据库性能调优：见 `database-tuning.md`
- 运维部署流程：见 `runtime-agent-deployment.md`
- 架构测试守卫：见 `code/backend/tests/README.md`

## 8. Guardrail

- Migration 文件一旦合并到 `main` 不得修改序号
- 每个 `.up.sql` 必须有对应 `.down.sql`
- 新增 migration 必须本地测试 `up` 和 `down`
- 生产环境执行 migration 前必须备份
- 删除列或表的 migration 必须在 commit message 中说明数据迁移策略

## 9. 已知限制

- 当前没有自动化的 migration 冲突检测
- 生产环境 migration 执行仍需手动触发
- 长时间锁表的 migration 需要业务方配合停机窗口
