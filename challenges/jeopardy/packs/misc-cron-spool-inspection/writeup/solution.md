# 解法

先看 `var/spool/cron/root`，能直接定位实际执行的脚本。再读脚本内容，就能在 `OUTPUT=` 变量里看到带 flag 名称的目标路径。
