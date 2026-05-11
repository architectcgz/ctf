# 解法

sqlite 在 WAL 模式下，最近的提交可能只落在 `-wal` 里。把主库和 sidecar 文件放在同目录后直接用 sqlite 打开查询，就能看到 WAL 中的记录。
