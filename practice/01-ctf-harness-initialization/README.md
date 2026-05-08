# CTF Harness Initialization

## 目标

严格参考 `deusyu/harness-engineering`，为 `ctf` 建立顶层 harness 结构。

## 方法

- 创建 `concepts/ thinking/ practice/ feedback/ works/ prompts/ references/`。
- 为每个目录创建 `AGENTS.md`。
- 创建 `scripts/check-consistency.sh`。
- 接入 `.githooks/pre-commit`。

## 验证

```bash
bash scripts/check-consistency.sh
```
