# AWD Checker Runner 运维说明

## 审计结果

AWD checker 的执行摘要写入 `check_result`。

- `script_checker` 和 `tcp_standard` 的每个 target 会带 `audit` 字段。
- `audit` 包含 `contest_id`、`service_id`、`team_id`、`round_number`、`checker_type`、`duration_ms`、`error_code`。
- `script_checker` 额外记录 `artifact_digest`、`exit_code`、截断后的 `stdout` / `stderr`。
- 持久化前会替换当前轮 flag，避免 checker 输出把真实 flag 写进结果。

## Artifact 目录

默认目录是：

```bash
./data/awd-checker-artifacts
```

可以通过环境变量覆盖：

```bash
AWD_CHECKER_ARTIFACT_DIR=/data/ctf/awd-checker-artifacts
```

目录结构：

```text
<artifact-root>/<challenge-slug>/<artifact-digest>/<package-relative-file>
```

同一 slug 重新导入脚本 checker 且 digest 变化时，平台会在导入事务成功后清理旧 digest 目录。事务失败不会清理旧目录。

## 故障定位

- `checker_artifact_unavailable`：artifact 文件缺失、越界、sha256 或 size 不匹配。
- `checker_runner_error`：runner 调用失败，通常检查 Docker runner、沙箱配置和日志。
- `checker_timeout` / `checker_output_limit_exceeded`：脚本运行超时或输出超限。
- `tcp_connect_failed`：TCP 目标不可达或地址模板错误。
- `tcp_step_failed`：TCP 步骤发送、读取或断言失败。
