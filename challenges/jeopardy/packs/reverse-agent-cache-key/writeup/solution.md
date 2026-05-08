# Reverse 题解：Agent Cache Keygen

## 解法

页面提供的是 Python 3.12 的 `pyc` 字节码，可以直接用标准库 `marshal` 读出 code object，再从常量里恢复兑换逻辑。

下面这段脚本会跳过 `pyc` 头，提取编码数组并恢复兑换码：

```bash
python3 - <<'PY'
import marshal
from pathlib import Path

pyc = Path("cache_gate.pyc").read_bytes()
code = marshal.loads(pyc[16:])

encoded = None
xor_key = None
prefix = None
for item in code.co_consts:
    if isinstance(item, tuple) and all(isinstance(v, int) for v in item):
        encoded = item
    elif isinstance(item, int):
        xor_key = item
    elif isinstance(item, str) and item.startswith("agent-"):
        prefix = item

decoded = "".join(chr(v ^ xor_key) for v in encoded)
print(prefix + decoded)
PY
```

得到兑换码：

```text
agent-cache-2026-key
```

然后访问兑换接口：

```bash
curl 'http://127.0.0.1:18081/redeem?code=agent-cache-2026-key'
```

返回 JSON 中的 `flag` 即为答案。
