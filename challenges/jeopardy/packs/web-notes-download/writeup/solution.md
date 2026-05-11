# Web 题解：内部笔记下载器

## 解法

下载接口把用户输入直接拼到 `NOTES_DIR / file_name`，却没有做路径规范化或目录限制检查，因此可以直接利用路径穿越：

```text
/download?file=../runtime/flag.txt
```

服务会先在 `runtime/flag.txt` 写入动态 Flag，再把这个路径当普通文件返回，所以访问上面的地址就能直接读到 Flag。
