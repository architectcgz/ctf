# 解法

镜像后续层虽然删除了 `.env`，但早期 layer 里仍保留完整文件内容。看 `manifest.json` 的 layer 顺序，再展开第一层就能直接看到 flag。
