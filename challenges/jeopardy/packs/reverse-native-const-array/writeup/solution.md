# 解法

检查逻辑本质是 `input[i] + i == encoded[i]`。常量表可从 rodata 或题包附带的辅助转储里提出来，对每个位置减去索引即可恢复原始 flag。
