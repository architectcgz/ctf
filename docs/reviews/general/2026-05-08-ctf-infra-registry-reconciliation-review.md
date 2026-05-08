# 2026-05-08 CTF Infra Registry Reconciliation Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/fix/docker-project-grouping`
- Branch: `fix/docker-project-grouping`
- Diff basis: `main..fix/docker-project-grouping`
- Runtime verification target:
  - `docker/ctf/docker-compose.dev.yml`
  - `docker/ctf/infra/**`
  - `scripts/registry/**`
  - `challenges/jeopardy/packs/*/challenge.yml`
  - `images / challenges / awd_challenges / contest_awd_services`

## Classification Check

- 认同 `development-pipeline` 的 `non-trivial / high-risk` 分类。
- 风险面集中在：
  - registry 配置事实源迁移
  - 运行态目录迁移
  - 题目镜像批量回灌
  - 数据库批量同步

## Gate Verdict

- `pass with minor issues`

## Findings

- 无 blocker 级发现。

## Material Findings

- 无。

## Senior Implementation Assessment

- 当前实现采用“保留单 Compose 入口 + 收口 infra 目录 + 复用现有 build 脚本 + 增补 DB 同步脚本”的路线，复杂度和风险都明显低于“重做导入链路”或“强行迁移 PostgreSQL/Redis 卷到 bind mount”。
- `sync-challenge-registry-state.py` 把 AWD 运行态里最容易漏掉的 `contest_awd_services.runtime_config` 与 `service_snapshot.runtime_config` 一并更新，这一点是必要的；否则数据库会出现题目主记录和比赛服务快照不一致。

## Required Re-validation

已完成：

```bash
bash -n scripts/registry/deploy-private-registry.sh
bash -n scripts/registry/build-and-push-challenge-image.sh
python3 -m py_compile scripts/registry/sync-challenge-registry-state.py
docker compose --profile registry -f docker/ctf/docker-compose.dev.yml config
mkdir -p backups && docker exec ctf-postgres pg_dump -U postgres -d ctf > backups/ctf-registry-sync-20260508.sql
bash scripts/registry/deploy-private-registry.sh --force-recreate
bash scripts/registry/build-and-push-challenge-image.sh <9 个容器题包>
python3 scripts/registry/sync-challenge-registry-state.py <9 个容器题包>
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id, name, tag, status, digest, source_type FROM public.images ORDER BY id;"
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id, package_slug, image_id FROM public.challenges ORDER BY id;"
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id, slug, runtime_config::jsonb->>'image_ref', runtime_config::jsonb->>'image_id' FROM public.awd_challenges ORDER BY id;"
docker exec ctf-postgres psql -U postgres -d ctf -P pager=off -c "SELECT id, awd_challenge_id, runtime_config::jsonb#>>'{challenge_runtime,image_ref}', runtime_config::jsonb#>>'{challenge_runtime,image_id}' FROM public.contest_awd_services ORDER BY id;"
git diff --check
```

## Residual Risk

- 这次 review 是同上下文自审，不是独立 reviewer gate。当前会话规则不允许在用户未明确要求 delegation 的情况下起 subagent，因此独立 review 仍未满足。
- 现有已经运行中的 `ctf-instance-*` / `ctf-workspace-*` 容器还是旧标签，只有后续新创建或重建的实例才会显示到 `ctf/jeopardy`、`ctf/awd` 分组下。
- `legacy-awd-2` 与 `e2e-script-checker-files` 仍保留历史镜像引用。这次没有删，也没有伪造它们“已重建成功”的状态。

## Touched Known-Debt Status

- 本次 touched surface 没有继续扩散到新的结构债点。
- 已收口的已知问题：
  - registry env 双份事实源
  - AWD build 脚本不支持 `docker/runtime/Dockerfile`
  - 比赛服务快照未跟随镜像回灌一起更新的风险
