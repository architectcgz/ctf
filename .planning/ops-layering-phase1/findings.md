# Findings

- `ops` 在 phase1 之后已经是 owner，但文件物理位置仍然是大平铺，后续继续迁移 `notification` 会放大维护成本。
- `identity`、`practice_readmodel` 已经提供了更合适的参考结构：根包只保留 contract，具体实现进入 `api/http`、`application`、`infrastructure`。
- `notification` phase2 与 `ops` 物理分层有明显耦合，最小总成本的做法是同轮完成。
