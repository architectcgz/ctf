# Misc 入门：注释里的便签题解

## 解法

查看 HTML 源码，可以看到注释：

```html
<!-- backup-note: ZmxhZ3ttaXNjX2NvbW1lbnRfbm90ZX0= -->
```

这是一段 Base64：

```bash
echo 'ZmxhZ3ttaXNjX2NvbW1lbnRfbm90ZX0=' | base64 -d
```

得到 flag：

```text
flag{misc_comment_note}
```
