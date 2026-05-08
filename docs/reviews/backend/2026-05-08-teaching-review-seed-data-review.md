# 2026-05-08 教学复盘样本数据 Review

## Scope

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- `docs/plan/impl-plan/2026-05-08-teaching-review-seed-data-implementation-plan.md`

## Findings

无 blocker。

## Checked Points

- 推荐查询已从 `LEFT JOIN + DISTINCT` 收口为 `EXISTS`，避免 PostgreSQL 下 `SELECT DISTINCT ... ORDER BY ...` 非法报错，同时不再因为标签 join 产生重复行。
- 新增测试覆盖了推荐链路的两个关键约束：
  - 已解题要被排除
  - 非同分类但命中知识点标签的题目仍可进入推荐结果
- 样本数据命令把写入范围限制在独立班级 `信安2401`，重复执行前会清理该批样本账号的 `instances / audit_logs / submissions / submission_writeups / skill_profiles / reports`，幂等边界明确。
- 命令内的开发环境默认值与 `scripts/dev-run.sh` 的本地依赖约定保持一致，避免再次误连 `shared-postgres:5432`。

## Residual Risks

- `seed-teaching-review-data` 当前是开发环境专用命令，默认会注入本地 `15432/16379` 连接参数；若在非本机 dev 环境执行，应显式覆盖相应 `CTF_POSTGRES_* / CTF_REDIS_*` 变量。
- 这次没有补 HTTP 端到端验证；当前证据来自命令内部直接调用 `assessment` 与 `teaching_readmodel` 服务，以及数据库落库结果。

## Validation Evidence

- `go test ./internal/module/challenge/infrastructure ./internal/module/assessment/... ./internal/module/teaching_readmodel/... ./cmd/seed-teaching-review-data`
- `go run ./cmd/seed-teaching-review-data`
- `docker exec ctf-postgres psql -U postgres -d ctf ...` 查询样本班级、事件数量和弱项维度
