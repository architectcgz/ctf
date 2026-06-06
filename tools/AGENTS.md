# Tools Directory

- `tools/` 存放项目工程工具入口，不承担 hook、README、AGENTS 和 harness 治理默认直接调用的稳定入口职责。
- 这里的命令可以被 `scripts/`、`harness/checks/`、文档流程或操作者显式调用，但默认不作为提交门禁的公开入口命名空间。
- 新增工具前，先判断它是不是应该成为稳定入口：
  - 如果主要面向 hook、review、doctor、workflow、README 常驻命令，放回 `scripts/` 顶层并纳入稳定入口清单。
  - 如果主要是工程辅助、同步、依赖准备、端到端演练或一次命令型工具，优先放在 `tools/`。
- `tools/` 下的可执行文件必须同步登记到 `harness/policies/script-layer-manifest.json`，由脚本层治理检查机械校验。
