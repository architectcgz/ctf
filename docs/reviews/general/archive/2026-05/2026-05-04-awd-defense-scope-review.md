# AWD 防守范围改造 Review

## Review Target

- Repository: `ctf`
- Diff source: 本地暂存前 AWD 防守页、后端 workspace API、AWD 上传解析、题库结构和文档改动
- Files reviewed: `code/frontend/src/components/contests/awd/*`、`code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`、`code/backend/internal/module/challenge/domain/awd_package_parser.go`、`code/backend/internal/module/contest/*`、`challenges/awd/**`

## Classification Check

同意按结构性改动处理。该改动同时影响前端展示、后端 DTO/API、上传预览校验、题库包结构和开发库快照，必须经过验证与 review gate。

## Gate Verdict

Pass with minor residual risk.

## Findings

- No material blocker found. 学生侧展示已从漏洞提示改为 `defense_scope` 边界信息；四道 AWD 题的 `editable_paths` 均为 `docker/challenge_app.py`，`docker/app.py` 均位于 `protected_paths`。
- Residual risk: Docker build 与 push registry 尚未完成端到端验证。此前 build 失败点为 Docker Hub 拉取 `python:3.12-slim` 的 TLS handshake timeout，属于外部 registry/network 依赖，需网络恢复后补跑。

## Senior Implementation Assessment

当前实现方向符合 AWD 训练目标：平台只声明可编辑范围、受保护运行契约和服务可用性契约，不把漏洞位置、修复步骤或攻击 payload 暴露到学生页。题库统一为 `docker/app.py` 固定入口、`docker/ctf_runtime.py` 平台契约、`docker/challenge_app.py` 业务漏洞代码，比 Web/TCP 分叉结构更容易被上传校验和前端展示一致处理。

## Required Re-validation

已执行：

```bash
go test ./internal/module/challenge/domain ./internal/module/challenge/application/commands ./internal/module/contest/application/queries
npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.test.ts
npm run typecheck
npm run check:theme-tail
npx eslint src/api/contracts.ts src/features/contest-awd-workspace/model/awdDefensePresentation.ts src/components/contests/awd/AWDDefenseOperationsPanel.vue src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts --quiet
git diff --check
```

待补充：

```bash
docker build -q -t ctf/awd-campus-drive:latest challenges/awd/ctf-1/awd-campus-drive/docker
docker build -q -t ctf/awd-iot-hub:latest challenges/awd/ctf-1/awd-iot-hub/docker
docker build -q -t ctf/awd-supply-ticket:latest challenges/awd/ctf-1/awd-supply-ticket/docker
docker build -q -t ctf/awd-tcp-length-gate:latest challenges/awd/ctf-1/awd-tcp-length-gate/docker
```

## Residual Risk

- 当前 review 未验证 registry push。
- 开发库已被批量更新过，但本 review 只覆盖代码和题库文件，不覆盖数据库运行态快照的再次查询证据。
