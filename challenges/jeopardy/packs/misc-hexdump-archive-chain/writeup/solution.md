# 解法

先把 hexdump 还原为原始字节，再根据 `file` / 魔数判断是 `bzip2 -> gzip -> tar.gz` 这条链。逐层拆开后就能读到 `flag.txt`。
