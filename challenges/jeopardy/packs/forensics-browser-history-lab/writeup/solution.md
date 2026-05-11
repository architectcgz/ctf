# 解法

浏览器历史本质上就是 sqlite。先看 `sqlite_master` 或直接查询 `urls` 表，再从 URL / title 中筛关键记录，能直接定位含 flag 的访问地址。
