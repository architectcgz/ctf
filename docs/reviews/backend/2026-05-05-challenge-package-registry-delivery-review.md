# 题包 Registry 交付闭环 Review

日期：2026-05-05

范围：题包导入、镜像构建任务、registry 校验、AWD readiness、管理端镜像状态展示。

## 结论

第一轮独立 review 发现 2 个 P1，均已修复并复测通过。

## Findings 与修复

### P1: Docker builder 未显式使用 registry 凭据

问题：

- `RegistryClient` 使用了 `container.registry` 凭据做 manifest 校验。
- `docker push` / `docker pull` 仍依赖宿主机预登录，导致平台配置了 registry 凭据但 worker 仍可能推送或拉取失败。

修复：

- `DockerCLIImageBuilder` 增加 `DockerCLIImageBuilderConfig`。
- composition 从 `container.registry` 传入 `server`、`username/password` 或 `identity_token`。
- `push` / `pull` 使用临时 `DOCKER_CONFIG` 写入认证配置，不执行全局 `docker login`，不污染宿主机 Docker 配置。
- 密钥不进入命令参数；错误信息只包含 `docker push/pull <ref>`。

覆盖测试：

- `TestDockerCLIImageBuilderPushUsesIsolatedDockerConfig`
- `TestDockerCLIImageBuilderPullUsesIsolatedDockerConfigForIdentityToken`

### P1: AWD readiness 可被 explicit access URL 绕过镜像状态

问题：

- AWD readiness 只检查 checker 配置和 preview 状态，没有把绑定镜像状态纳入硬门禁。
- `PreviewChecker` 填写 `AccessURL` 时可绕过自动拉起镜像路径，即使题目绑定的 `image_id` 仍处于 `pending/failed`。

修复：

- `AWDReadinessChallengeRecord` / domain readiness 增加 runtime image ID/status。
- readiness 从 `contest_awd_services.runtime_config` 和服务 snapshot 解析 `image_id`，批量读取 `images.status`。
- 绑定镜像非 `available` 时返回 `image_not_available`，并阻断 AWD 开赛和自动状态推进。
- explicit `AccessURL` 路径只在解析到绑定 `image_id` 时检查镜像状态；旧的纯手动 URL 预检仍兼容。

覆盖测试：

- `TestAWDQueryServiceGetReadinessBlocksUnavailableRuntimeImage`
- `TestAWDServicePreviewCheckerRejectsExplicitAccessURLWhenRuntimeImageUnavailable`

## Review 清单

- 导入路径不再直接把平台构建镜像标为 `available`。
- 平台构建模式不要求上传者填写完整镜像名，只接受 tag 建议。
- 外部镜像引用路径必须通过 manifest / pull / inspect 后才可用。
- build / push / verify 失败会写入失败摘要并阻断发布或 AWD readiness。
- registry 凭据不进入题包、API response 或 docker 命令参数。
- worker 使用应用 lifecycle context 启停；本地 registry e2e 临时容器已清理。

## 复测命令

```bash
cd code/backend
go test ./internal/module/challenge/application/commands -run 'Test(ImageBuildService|CommitChallengeImport|AWDChallengeImport|DockerCLIImageBuilder)' -count=1 -timeout=120s
go test ./internal/module/contest/application/queries -run 'TestAWDQueryServiceGetReadiness' -count=1 -timeout=120s
go test ./internal/module/contest/application/commands -run 'AWDService|Readiness|Image|PreviewChecker' -count=1 -timeout=120s
go test ./internal/module/practice/application/commands -run 'TestServiceStartContestAWDServiceCanProvisionFromContestAWDServiceSnapshot' -count=1 -timeout=120s
```

结果：全部通过。

## 残余风险

- 真正的私有 registry 权限、insecure registry 和 TLS 证书组合仍需要部署环境 smoke test。
- 当前第一阶段只覆盖单镜像题包；拓扑多节点镜像构建仍按设计文档作为后续扩展。
