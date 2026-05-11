# 解法

接口会识别 `<!ENTITY x SYSTEM "file://...">` 并把 `&x;` 替换成本地文件内容，因此构造一个指向 `flag.txt` 的实体即可读到 flag。
