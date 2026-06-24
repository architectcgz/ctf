# 数据库连接池与性能调优

> 状态：Current
> 事实源：`code/backend/internal/infrastructure/persistence/postgres.go`、生产环境配置
> 替代：无

## 本文档范围

| 覆盖 | 不覆盖 |
|------|--------|
| PostgreSQL 连接池配置参数 | PostgreSQL 安装与初始化 |
| 慢查询监控策略 | 完整的数据库备份方案 |
| 索引策略建议 | 具体表的索引实现（见各模块文档） |
| 性能调优检查清单 | 生产环境硬件配置选型 |

## 定位

本文档只说明后端数据库连接池配置、慢查询监控、索引策略和性能调优的关键检查点。

## 当前设计

- `code/backend/internal/infrastructure/persistence/postgres.go`
  - 负责：初始化 PostgreSQL 连接池，配置超时、最大连接数、空闲连接数等参数
  - 不负责：具体业务查询优化或表结构设计

## 1. 连接池配置

### 1.1 关键参数

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| `MaxOpenConns` | `25` | 最大打开连接数，避免耗尽数据库连接 |
| `MaxIdleConns` | `10` | 最大空闲连接数，减少频繁建连开销 |
| `ConnMaxLifetime` | `5min` | 连接最大生命周期，避免长连接积累问题 |
| `ConnMaxIdleTime` | `2min` | 空闲连接最大保持时间 |

### 1.2 配置示例

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(2 * time.Minute)
```

### 1.3 调优建议

- **CPU 密集型应用**：适当降低连接数，避免上下文切换开销
- **IO 密集型应用**：适当提高连接数，提升并发吞吐
- **连接池监控**：定期检查 `db.Stats()` 中的 `WaitCount` 和 `WaitDuration`
- **连接泄漏检测**：如果 `InUse` 持续增长不下降，说明存在连接泄漏

## 2. 慢查询监控

### 2.1 PostgreSQL 配置

在 `postgresql.conf` 中启用慢查询日志：

```
log_min_duration_statement = 1000  # 记录超过 1 秒的查询
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d '
log_statement = 'none'
```

### 2.2 监控策略

- **定期检查慢查询日志**：`grep "duration:" postgresql.log`
- **分析 EXPLAIN ANALYZE**：对慢查询执行计划分析
- **监控指标**：
  - 平均查询时间
  - P95 / P99 查询时间
  - 全表扫描次数
  - 索引命中率

### 2.3 常见慢查询场景

| 场景 | 原因 | 优化方向 |
|------|------|---------|
| 大表全表扫描 | 缺少索引 | 添加合适的索引 |
| `JOIN` 慢 | 关联字段无索引 | 为外键添加索引 |
| `ORDER BY` 慢 | 排序字段无索引 | 添加排序字段索引 |
| `COUNT(*)` 慢 | 大表统计 | 考虑缓存或估算 |
| `LIKE '%keyword%'` | 前缀通配符 | 改用全文检索或 trigram 索引 |

## 3. 索引策略

### 3.1 索引创建原则

- **高频查询字段优先**：`WHERE`、`JOIN`、`ORDER BY` 中的字段
- **选择性高的字段优先**：区分度高的字段索引效果更好
- **复合索引顺序**：最常用的过滤条件放在最前面
- **避免过度索引**：每个索引都会增加写入开销

### 3.2 常见索引类型

| 索引类型 | 适用场景 | 示例 |
|---------|---------|------|
| B-tree | 等值查询、范围查询 | `CREATE INDEX idx_user_id ON challenges(user_id)` |
| 唯一索引 | 保证唯一性 | `CREATE UNIQUE INDEX idx_username ON users(username)` |
| 复合索引 | 多字段联合查询 | `CREATE INDEX idx_contest_user ON submissions(contest_id, user_id)` |
| 部分索引 | 条件过滤 | `CREATE INDEX idx_active_users ON users(id) WHERE deleted_at IS NULL` |
| GIN | 数组、JSONB、全文检索 | `CREATE INDEX idx_tags ON challenges USING GIN(tags)` |

### 3.3 索引命名约定

```
idx_<表名>_<字段名>
idx_<表名>_<字段1>_<字段2>  -- 复合索引
uidx_<表名>_<字段名>         -- 唯一索引
```

### 3.4 索引维护

```sql
-- 查看表的索引使用情况
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan;

-- 查找从未使用的索引
SELECT
    schemaname,
    tablename,
    indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE 'pg_toast%';

-- 重建索引（生产环境谨慎操作）
REINDEX INDEX CONCURRENTLY idx_name;
```

## 4. 性能调优检查清单

### 4.1 开发阶段

- [ ] 新增查询是否有对应索引
- [ ] 是否避免了 `SELECT *`
- [ ] 是否使用了参数化查询（防 SQL 注入）
- [ ] 是否避免了 `N+1` 查询问题
- [ ] 大表分页是否使用了游标或 `LIMIT + OFFSET` 优化

### 4.2 测试阶段

- [ ] 慢查询日志是否开启
- [ ] 是否对关键查询执行了 `EXPLAIN ANALYZE`
- [ ] 是否测试了大数据集场景（1万+ 行）
- [ ] 连接池配置是否合理
- [ ] 是否有数据库连接泄漏

### 4.3 生产环境

- [ ] 定期检查慢查询日志
- [ ] 监控数据库连接数
- [ ] 监控索引命中率
- [ ] 定期清理无用索引
- [ ] 定期执行 `VACUUM` 和 `ANALYZE`

## 5. 常见性能陷阱

### 5.1 N+1 查询

**问题**：

```go
// 错误示例：查询 10 个比赛后，再逐个查询每个比赛的题目
contests := getContests()
for _, c := range contests {
    challenges := getChallengesByContestID(c.ID)  // N+1 查询
}
```

**解决**：

```go
// 正确示例：一次性查询所有题目
contests := getContests()
contestIDs := extractIDs(contests)
challenges := getChallengesByContestIDs(contestIDs)  // 批量查询
```

### 5.2 过度使用 `DISTINCT`

**问题**：`DISTINCT` 需要排序和去重，开销较大

**解决**：优先使用 `GROUP BY` 或在业务层去重

### 5.3 大表 `OFFSET` 深分页

**问题**：`OFFSET 10000` 会扫描前 10000 行后丢弃

**解决**：使用游标分页（Cursor-based pagination）

```sql
-- 游标分页示例
SELECT * FROM submissions
WHERE id > $last_seen_id
ORDER BY id
LIMIT 20;
```

### 5.4 隐式类型转换

**问题**：

```sql
-- user_id 是 BIGINT，但传入了字符串
SELECT * FROM users WHERE user_id = '123';  -- 索引失效
```

**解决**：确保查询参数类型与字段类型一致

## 6. 连接池监控指标

### 6.1 Go `sql.DB` 统计

```go
stats := db.Stats()
log.Printf("Open connections: %d", stats.OpenConnections)
log.Printf("In use: %d", stats.InUse)
log.Printf("Idle: %d", stats.Idle)
log.Printf("Wait count: %d", stats.WaitCount)
log.Printf("Wait duration: %s", stats.WaitDuration)
```

### 6.2 告警阈值建议

| 指标 | 阈值 | 说明 |
|------|------|------|
| `WaitCount` | > 100/min | 连接池不足，考虑增加连接数 |
| `WaitDuration` | > 100ms | 等待连接时间过长 |
| `InUse` / `MaxOpenConns` | > 80% | 连接池接近饱和 |
| `Idle` | 长期为 0 | 连接池配置可能过小 |

## 7. 边界

### 7.1 本文档不覆盖

- **数据库备份与恢复**：见运维手册
- **高可用方案**：主从复制、故障切换
- **分库分表**：当前单库单表即可满足需求
- **读写分离**：当前未实施

### 7.2 与其他文档的关系

- Migration 管理：见 `database-migration.md`
- 后端模块设计：见 `docs/architecture/backend/modules/`
- 测试策略：见 `code/backend/tests/README.md`

## 8. Guardrail

- 新增查询必须经过 `EXPLAIN ANALYZE` 验证
- 大表操作必须在测试环境验证性能
- 生产环境修改索引前必须备份
- 连接池配置变更必须先在测试环境验证

## 9. 已知限制

- 当前没有自动化的慢查询告警
- 连接池监控指标未接入可观测性平台
- 索引命中率监控依赖手动执行 SQL 查询
