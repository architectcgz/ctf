# 题包拓扑同步与导出设计

## 目标

让标准题包通过 `docker/topology.yml` 描述拓扑，并在导入、编辑和导出过程中保留题包来源、拓扑基线、源码树和修订记录。

这份文档承接 `docs/superpowers/plans/2026-04-21-package-topology-sync.md` 的已实现结果。当前以本文和代码为最终事实源，原 plan 只作为实施过程记录。

## 当前状态

- `challenge.yml` 可通过 `extensions.topology.source` 指向题包内拓扑文件。
- 导入预览会解析拓扑摘要和题包文件树。
- 导入提交会保存 `challenge_topologies` 和 `challenge_package_revisions`。
- 拓扑工作台能展示题包来源、基线、漂移状态、文件树、修订历史和导出入口。
- 题包导出会基于当前挑战与拓扑状态生成完整 zip。

## 设计原则

### 1. 题包拓扑是 package-native 输入

题包内使用 `docker/topology.yml` 作为拓扑来源。平台导入时把它映射成内部 `TopologySpec`，但不把内部结构反向要求给题包作者。

入口字段：

```yaml
extensions:
  topology:
    enabled: true
    source: docker/topology.yml
```

`source` 必须是题包内部相对路径，不能越过题包根目录。

### 2. 平台保存 provenance

导入后，拓扑记录需要保留来源信息：

- `source_type = package_import`
- `source_path`
- `package_revision_id`
- `package_baseline_spec`
- `sync_status`
- `last_export_revision_id`

这些字段用于区分“平台手动创建的拓扑”和“来自题包的拓扑”，也用于判断当前拓扑是否已经偏离导入基线。

### 3. 修订记录保存完整题包上下文

`challenge_package_revisions` 保存每次导入或导出的题包修订：

- `revision_no`
- `source_type`
- `parent_revision_id`
- `package_slug`
- `archive_path`
- `source_dir`
- `manifest_snapshot`
- `topology_source_path`
- `topology_snapshot`

导入修订用于追溯原始题包；导出修订用于下载和后续再导入。

## 题包拓扑格式

第一版支持：

- `api_version: v1`
- `kind: topology`
- `entry_node_key`
- `networks`
- `nodes`
- `links`
- `policies`

节点至少需要：

- `key`
- `image.ref`
- 入口节点的 `service_port`

如果没有显式网络，平台会补默认网络。若没有节点标记 `inject_flag`，平台会把入口节点作为默认 flag 注入点。

## 导入流程

1. 上传题包。
2. 解析 `challenge.yml`。
3. 如果 `extensions.topology.enabled = true`，读取 `extensions.topology.source`。
4. 解析 package-native topology。
5. 生成导入预览，包含拓扑摘要和题包文件树。
6. 提交导入后创建或更新 challenge。
7. 保存完整题包源码树和导入 revision。
8. 将 package topology 映射为平台 `challenge_topologies`。

## 编辑与漂移

拓扑工作台可以在导入后继续编辑平台拓扑。当前漂移规则：

- 当前拓扑与导入基线一致：`sync_status = clean`。
- 当前拓扑偏离导入基线：`sync_status = drifted`。

漂移状态只表示平台拓扑与最近题包基线不同，不表示配置错误。

## 导出流程

导出以当前平台状态为准：

1. 读取 challenge、writeup、hint、runtime 和当前 topology。
2. 以最近 package revision 的源码树为基础。
3. 重写 `challenge.yml`。
4. 重写 `docker/topology.yml`。
5. 生成新的 exported revision。
6. 更新 topology 的 `last_export_revision_id`。
7. 返回下载地址。

导出接口：

- `POST /api/v1/authoring/challenges/:id/package-export`
- `GET /api/v1/authoring/challenges/:id/package-export/download?revision_id=:rid`

## 前端落点

导入预览：

- 展示拓扑入口、节点、网络、策略摘要。
- 展示题包文件树。
- 提示导入会保存源码树和拓扑基线。

拓扑工作台：

- 展示来源路径和同步状态。
- 展示导入基线、当前状态和修订历史。
- 提供导出按钮。
- 保留模板库和平台手动拓扑能力。

## 风险与约束

- 不要把题包 `docker/topology.yml` 直接当作平台内部 JSONB 结构；必须经过解析和映射。
- 不要丢弃题包源码树，否则无法可靠导出完整包。
- 不要把 `drifted` 当成错误；它只是表示平台状态已经不同于导入基线。
- 导出应重写 manifest 和 topology，不能只把原始 zip 原样返回。

## 验收标准

- 导入预览能返回拓扑摘要和题包文件树。
- 导入提交能创建 `challenge_topologies` 和 package revision。
- 拓扑保存保留 package provenance 并更新漂移状态。
- 导出能生成包含 `challenge.yml` 和 `docker/topology.yml` 的完整题包。
