# Web 题解：双层伪装

## 解法

页面前面几段注释和变量都是干扰项，真正的线索是最后一个变量：

```javascript
var _0x4a3f = "Wm14aFozdDNaV0l0YzI5MWNtTmxMV0YxWkdsMExXUnZkV0pzWlMxM2NtRndMVEF4ZlE9PQ==";
```

这段值先解一次 Base64，会得到另一段仍然像 Base64 的文本；再解一次，就能恢复真正的 flag：

```bash
python3 - <<'PY'
import base64
s = "Wm14aFozdDNaV0l0YzI5MWNtTmxMV0YxWkdsMExXUnZkV0pzWlMxM2NtRndMVEF4ZlE9PQ=="
print(base64.b64decode(base64.b64decode(s)).decode())
PY
```
