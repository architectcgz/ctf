# 解法

PDF 可以内嵌文件或直接把数据放进 stream。查对象结构后定位 `EmbeddedFile` 对象，读出对应 stream 就能拿到 flag。
