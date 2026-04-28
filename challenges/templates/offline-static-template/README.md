# 离线题模板说明

这个模板适合：

- 编码题
- 基础 crypto
- 基础 reverse
- 静态取证样本

复制后你需要做的事情：

1. 改 `challenge.yml`
2. 改 `statement.md`
3. 用真实样本替换 `attachments/challenge.txt`
4. 重新计算附件 `sha256`
5. 刷新 `dist/<slug>.zip`

## 重新计算 sha256

```bash
sha256sum attachments/challenge.txt
```

把输出结果填回 `challenge.yml` 的 `attachments[].sha256`。

## 自查清单

- 附件不是空文件
- 附件不是占位文本
- 题面和附件一致
- 老师自己能从附件恢复答案
