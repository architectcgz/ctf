# 解法

用邮件解析器直接读 eml，遍历 multipart 的每个 part。正文只是提示，真正的值在附件 `note.txt` 里，解码后即可得到 flag。
