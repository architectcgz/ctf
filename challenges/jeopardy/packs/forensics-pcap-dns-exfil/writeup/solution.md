# 解法

抓包里每个 DNS Query Name 都带了一个编号和一段 Base32 数据。按编号排序后拼起来，再补齐 `=` 做 Base32 解码即可恢复 flag。
