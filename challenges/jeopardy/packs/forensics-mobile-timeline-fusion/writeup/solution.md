# 解法

关键线索在被删除的聊天记录里，正文直接给出了：

- `incident` 标记
- 起止时间窗口
- `assemble=utc`

拿到这条线索后，剩下就是做时间线拼接：

1. 从 `clipboard.json`、`sync.log`、`drafts.csv` 和 `notes.ndjson` 里筛出同一个 incident 的记录。
2. 把它们各自的时间统一转换成 UTC。
3. 按 UTC 排序后，分别按各自编码方式解码：
   - `base64`
   - `rot13`
   - `hex`
   - `reverse`
4. 拼接四段明文，得到完整 flag。

这题的难点不在单个文件格式，而在多源取证里先定窗口、再统一时区、最后做编码还原。
